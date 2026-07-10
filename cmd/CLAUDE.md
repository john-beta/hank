# CLAUDE.md

Project context for AI code generation. Follow these patterns when writing or modifying code.

---

## Project overview

A CodeAct-style AI agent built in Go. The agent uses executable code as its action mechanism: the LLM writes code, Go executes it, captures output/errors, and feeds them back so the LLM can self-correct. The system is phase-driven — each phase defines its own instructions, tools, and execution logic behind a shared interface.

**Module:** `github.com/john-beta/hank/cmd`

---

## Repository structure

```
cmd/
├── internal/
│   ├── agent/          ← orchestration layer (stateless, transport-agnostic)
│   │   ├── phases/     ← business logic layer (phase interface + implementations)
│   │   │   ├── phase/       ← shared Interface + ID contract
│   │   │   ├── phase_one/
│   │   │   └── phase_two/
│   │   ├── agent.go
│   │   ├── events.go
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
                   agent  →  phases (business logic)
```

**Core rules:**

- Dependencies point inward only. No layer imports a layer above it.
- `agent` depends on `llm` and `store` only through the interfaces those packages expose (`llm.Client`, `store.Store`). It never imports the OpenAI SDK or `database/sql` directly.
- `phases` depends on `llm` (for type definitions like `ToolDef`) and its own `phase` subpackage, which holds the shared `Interface` and `ID` contract. It never imports `agent`.
- `main.go` is the only place that constructs concrete implementations and wires them together.

### Layer responsibilities

**`transport/http`** — Translation only. Parses HTTP requests into DTOs, delegates to the agent, translates agent events into SSE format. Zero business logic, zero phase awareness, zero SQL.

**`agent`** — Orchestration. Owns the ReAct loop (`loop.go`), state management (`state.go`), persistence coordination (`persist.go`), and event emission (`events.go`). Stateless across requests: loads state from store at the start, persists changes at the end. Talks to phases only through the shared phase interface. Does not know what the phases do internally.

**`phases`** — Business logic. Each phase is a self-contained unit implementing the shared phase interface. A phase bundles: instructions (system prompt), tool definitions (what the LLM can call), and tool execution (what happens when the LLM calls a tool). Each phase lives in its own subpackage with its own files for instructions, tools, and execution. The registry (`registry.go`) maps phase IDs to concrete implementations. **Phase transitions are not yet implemented** — the phase is decided once per request from the persisted session state and held fixed for the whole loop; advancing a session to a new phase is future work, not something the interface currently expresses.

**`llm`** — OpenAI adapter. Wraps the SDK behind a `Client` interface. Translates between agent-level types (`Request`, `StreamEvent`) and SDK types (`ResponseNewParams`, stream events). The only package that imports `github.com/openai/openai-go/v3`.

**`store`** — SQLite adapter. Implements a `Store` interface with `database/sql`. Manages sessions and turns. The only package that imports the SQLite driver.

### Key patterns

**Phase-driven turns.** At the start of a request, the loop resolves the current phase once via the registry, then reads `Instructions()` and `Tools()` from it fresh on every iteration. Instructions and tools are NOT inherited between turns — they are explicitly re-sent each time. This prevents hallucination from stale context and enables each phase to fully control what the model sees.

**Streaming via channels.** `agent.Handle()` returns a `<-chan Event`. The loop goroutine writes events as they happen; the HTTP handler reads and translates them to SSE. The agent never writes HTTP directly.

**Context cancellation chain.** `r.Context()` propagates from HTTP client → agent loop goroutine → LLM stream goroutine. All goroutines respect `ctx.Done()` via `select`. When the client disconnects, everything tears down cleanly. No orphaned goroutines.

**Stateless agent.** The `Agent` struct has no per-session state. State is loaded from SQLite at the start of each request (`persist.go`) and written back at the end. A single `Agent` instance serves all sessions safely.

**Phase interface contract:**

```go
type Interface interface {
    Instructions() string
    Tools() []llm.ToolDef
    Execute(name, args string) (string, error)
}
```

Every phase implements these three methods. The loop only calls these — it never reaches into phase internals. There is no transition/`Next` method yet; phase advancement is a known gap, not a design choice to preserve.

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

SQLite with two tables:

- `session` — conversation container with current phase.
- `turn` — each user message or agent response, tagged with the phase active at creation time.

Tables created via `CREATE TABLE IF NOT EXISTS` on startup. No migration framework.

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

## Adding a new phase (checklist)

1. Create `internal/agent/phases/phase_NAME/` with three files: `phase.go`, `instructions.go`, `tools.go`.
2. Implement the phase interface (use `var _ phase.Interface = Phase{}` for a compile-time check).
3. Add an `ID` constant to the const block in `internal/agent/phases/phase/phase.go`.
4. Add one `case` in `phases/registry.go` → `Get()`.
5. Nothing else changes — the loop, persistence, and transport layers are unaware of the new phase.
