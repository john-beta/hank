# CLAUDE.md

Project context for AI code generation. Follow these patterns when writing or modifying code.

---

## Project overview

A CodeAct-style AI agent built in Go — a **Workspace Curator**. The agent uses executable code as its action mechanism: the LLM writes code, Go executes it, captures output/errors, and feeds them back so the LLM can self-correct. The system is **mode-driven**: there are exactly two modes, Planning and Executing, and the active mode is *derived* from a single domain boolean, `session.approved_proposal` (false → Planning, true → Executing). Each mode defines its own instructions, tools, and execution logic behind a shared interface.

**Module:** `github.com/john-beta/hank/cmd`

---

## Repository structure

```
cmd/
├── internal/
│   ├── agent/          ← orchestration layer (stateless, transport-agnostic)
│   │   ├── modes/      ← business logic layer (mode interface + implementations)
│   │   │   ├── mode/        ← shared Interface + ID contract + Resolve + AutoReFeed
│   │   │   ├── planning/    ← Planning mode (approved_proposal == false)
│   │   │   └── executing/   ← Executing mode (approved_proposal == true)
│   │   ├── agent.go
│   │   ├── events.go
│   │   ├── input.go        ← discriminated TurnInput (message | tool result)
│   │   ├── loop.go
│   │   ├── persist.go
│   │   └── state.go
│   ├── llm/            ← OpenAI SDK wrapper (only layer that knows OpenAI exists)
│   ├── store/          ← SQLite persistence (only layer that knows SQL exists)
│   └── transport/
│       └── http/       ← HTTP handlers, SSE streaming, DTOs
├── server/
│   ├── main.go         ← wiring: constructs all implementations, injects, starts server
│   └── agent.db
└── go.mod
```

---

## Backend (Go)

### Architecture: layered with dependency injection via interfaces

```
transport/http  →  agent  →  llm
                           →  store
                   agent  →  modes (business logic)
```

**Core rules:**

- Dependencies point inward only. No layer imports a layer above it.
- `agent` depends on `llm` and `store` only through the interfaces those packages expose (`llm.Client`, `store.Store`). It never imports the OpenAI SDK or `database/sql` directly.
- `modes` depends on `llm` (for type definitions like `ToolDef`) and its own `mode` subpackage, which holds the shared `Interface`, the `ID` contract, and the pure `Resolve`/`AutoReFeed` helpers. It never imports `agent`.
- `main.go` is the only place that constructs concrete implementations and wires them together.

### Layer responsibilities

**`transport/http`** — Translation only. Parses HTTP requests into DTOs (a discriminated `TurnInput`: plain message or tool result), delegates to the agent, translates agent events into SSE format. Zero business logic, zero mode awareness, zero SQL.

**`agent`** — Orchestration. Owns the ReAct loop (`loop.go`), state management (`state.go`), persistence coordination (`persist.go`), event emission (`events.go`), and the turn-input contract (`input.go`). Stateless across requests: loads state from store at the start, persists changes as it goes. Talks to modes only through the shared mode interface. Does not know what the modes do internally. The Planning → Executing transition is handled here, *before* the loop (`persist.go`): a tool result carrying `approved: true` flips `approved_proposal` in the store; the loop never transitions.

**`modes`** — Business logic. Each mode is a self-contained unit implementing the shared mode interface. A mode bundles: instructions (system prompt), tool definitions (what the LLM can call), and tool execution (what happens when the LLM calls a tool). Each mode lives in its own subpackage (`planning/`, `executing/`) with its own files for instructions, tools, and execution. The registry (`registry.go`) maps mode IDs to concrete implementations; `emptyMode` is the null object. The active mode is **derived, never persisted** — `mode.Resolve(approvedProposal)` computes it once per request. There is no transition/`Next` method: the Planning → Executing move is a one-way boolean flip, resolved pre-loop.

**`llm`** — OpenAI adapter. Wraps the SDK behind a `Client` interface. Translates between agent-level types (`Request`, `StreamEvent`) and SDK types (`ResponseNewParams`, stream events). Sets `ParallelToolCalls: false` (one function call per response). The only package that imports `github.com/openai/openai-go/v3`. `ToolDef.AutoReFeed` is agent metadata and is stripped here — it never reaches OpenAI.

**`store`** — SQLite adapter. Implements a `Store` interface with `database/sql`. Manages sessions, turns, and calls. The only package that imports the SQLite driver.

### Key patterns

**Mode-driven turns.** At the start of a request the mode is resolved once from `session.approved_proposal` via `mode.Resolve`, then held fixed for the whole loop. The loop reads `Instructions()` and `Tools()` from it fresh on every iteration. Instructions and tools are NOT inherited between turns — they are explicitly re-sent each time. This prevents hallucination from stale context and enables each mode to fully control what the model sees.

**`auto_re_feed` drives the loop.** Each tool declares a static `AutoReFeed` bool. When the model emits a call, the loop looks it up (`mode.AutoReFeed`): `true` → execute the tool, record the result, and re-feed automatically; `false` → emit the enriched `tool_call` SSE event (with `call_id` + `auto_re_feed`) and stop the turn so the client resolves it and sends the result back in the next request. Client-sourced tool results and loop-sourced auto results travel the exact same `llm.Request{ToolResults, PrevResponseID}` path — one Request type, two sources.

**One row per call.** A call is INSERTed once (result NULL) when emitted, then UPDATEd in place when it resolves. `PendingCall` finds the single unresolved non-auto call on a session's latest agent turn.

**Streaming via channels.** `agent.Handle()` returns a `<-chan Event`. The loop goroutine writes events as they happen; the HTTP handler reads and translates them to SSE. The agent never writes HTTP directly.

**Context cancellation chain.** `r.Context()` propagates from HTTP client → agent loop goroutine → LLM stream goroutine. All goroutines respect `ctx.Done()` via `select`. When the client disconnects, everything tears down cleanly. No orphaned goroutines.

**Stateless agent.** The `Agent` struct has no per-session state. State is loaded from SQLite at the start of each request (`persist.go`) and written back at the end. A single `Agent` instance serves all sessions safely.

**Mode interface contract:**

```go
type Interface interface {
    Instructions() string
    Tools() []llm.ToolDef
    Execute(name, args string) (string, error)
}
```

Every mode implements these three methods. The loop only calls these — it never reaches into mode internals. There is deliberately no transition/`Next` method: mode is derived from `approved_proposal`, and the Planning → Executing flip is resolved pre-loop, not on the interface.

### Conventions

- **Small files, one responsibility each.** No file should exceed ~200 lines. If it grows, split by sub-responsibility.
- **Interfaces defined by the provider package**, next to the implementation they gate (`llm.Client` in `llm`, `store.Store` in `store`). Consumers depend on the interface, not the concrete type.
- **Error messages identify the operation:** `"store: get session %s: %w"`, not `"database error"`.
- **No global state.** Everything flows through constructors and method parameters.
- **`internal/` enforces encapsulation.** Nothing outside `cmd/internal/` can import these packages.

### OpenAI SDK constraints

- **Responses API only** (`client.Responses.NewStreaming`, `client.Responses.New`). Never Chat Completions.
- **Import paths:** `github.com/openai/openai-go/v3`, `.../v3/responses`, `.../v3/option`.
- **Instructions do not carry over between turns** when using `PreviousResponseID`. Every turn must explicitly set its instructions and tools.
- **Tool results re-feeding:** when submitting tool outputs back, use `OfInputItemList` with `OfFunctionCallOutput`, include `PreviousResponseID`, and re-send instructions + tools.

### Database

SQLite with three tables:

- `session` — conversation container bound to a workspace: `root_dir` and the `approved_proposal` boolean (the only persisted mode signal).
- `turn` — each user message or agent response. `output_text` holds a user's message or an agent's streamed text; it is NULL for a user turn that carries a tool result (that payload lives in `call.result`).
- `call` — one row per tool call. INSERTed with `result` NULL when the agent emits the call, UPDATEd in place when it resolves. Carries `auto_re_feed` so pending client-resolved calls can be found.

There is no persisted `phase`/`mode` column — the mode is always derived from `approved_proposal`. Tables created via `CREATE TABLE IF NOT EXISTS` on startup. No migration framework.

---

## Development commands

The Go module root is `cmd/` (that's where `go.mod` lives) — run Go commands from inside `cmd/`, not the repo root.

```bash
# Run the server (from cmd/)
cd cmd
OPENAI_API_KEY=sk-... go run ./server

# Build
go build -o agent ./server

# Tidy dependencies
go mod tidy
```

---

## Modes (Planning & Executing)

There are exactly two modes; the model is *while `approved_proposal` is false the agent plans; once it is true the agent executes*. Each mode lives in `internal/agent/modes/<name>/` with three files — `mode.go`, `instructions.go`, `tools.go` — and implements the mode interface (`var _ mode.Interface = Mode{}` for a compile-time check). The registry (`modes/registry.go`) maps `mode.Planning`/`mode.Executing` IDs to the implementations.

The mode boundary is structural: `RunExecution` is present only in Executing's `Tools()` and dispatch; `ProposeStructure` is present only in Planning's. Adding a genuinely new mode would mean a new `ID` constant + `Resolve` branch, but the current design is fixed at two — do not add a third speculatively.

**Scaffold status:** tool handlers are stubs (`"not yet implemented"`) and both instruction strings are empty. The system prompts and tool execution logic are the intended next step, written on top of this scaffold.
