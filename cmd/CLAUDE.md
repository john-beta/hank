# CLAUDE.md

Project context for AI code generation. Follow these patterns when writing or modifying code.

---

## Project overview

A CodeAct-style AI agent built in Go — a **Workspace Curator**. The agent uses executable code as its action mechanism: the LLM writes code, Go executes it, captures output/errors, and feeds them back so the LLM can self-correct. The system is **mode-driven**: there are exactly two modes, Planning and Executing, and the active mode is *derived* from a single domain boolean, `ApprovedProposal` (false → Planning, true → Executing). Unlike the rest of a session's state, `ApprovedProposal` is **never persisted** — it lives only in memory for the duration of one request (see "Mode-driven turns" under [Key patterns](#key-patterns), and [Database](#database)). Each mode defines its own instructions, tools, and execution logic behind a shared interface.

**Module:** `github.com/john-beta/hank/cmd`

---

## Repository structure

```
cmd/
├── internal/
│   ├── agent/            ← orchestration layer (stateless, transport-agnostic)
│   │   ├── modes/            ← business logic layer (mode interface + implementations)
│   │   │   ├── mode/              ← shared Interface contract + AutoReFeed (no registry/ID — see below)
│   │   │   ├── planning/          ← Planning mode (ApprovedProposal == false)
│   │   │   └── executing/         ← Executing mode (ApprovedProposal == true)
│   │   ├── agent.go          ← Agent struct, New, Handle (public API) + run (its goroutine body)
│   │   ├── state.go          ← per-request State + StateFromStore
│   │   ├── input.go          ← TurnInput/ToolResultInput contract + how input becomes an llm.Request
│   │   ├── events.go         ← Event/EventType (SSE contract) + emit/fail
│   │   ├── loop.go           ← ReAct loop skeleton: maxIterations, runLoop
│   │   ├── step.go           ← one loop iteration: runStep, consume, streamResult, saveAgentTurn
│   │   └── call.go           ← one tool call: handleCall
│   ├── llm/              ← OpenAI SDK wrapper (only layer that knows OpenAI exists)
│   ├── store/            ← persistence boundary (only layer that knows SQL exists)
│   │   ├── interface.go      ← Store interface + Session/Turn/Call domain types
│   │   └── sqlite/           ← the (only) SQLite implementation, split by entity
│   │       ├── sqlite.go         ← SQLiteStore, NewSQLite, Close
│   │       ├── migrations.go     ← schema (session/turn/call tables), RunMigrations
│   │       ├── session.go        ← CreateSession, GetSession
│   │       ├── turn.go           ← SaveTurn, LastAgentTurn
│   │       └── call.go           ← SaveCall, UpdateCallResult, PendingCall
│   └── transport/
│       └── http/         ← HTTP handlers, SSE streaming, DTOs
├── server/
│   ├── main.go           ← wiring: constructs all implementations, injects, starts server
│   └── agent.db
└── go.mod
```

`cmd/internal/store` and `cmd/internal/store/sqlite` are two separate Go packages (different directories, different import paths) that both happen to declare `package store` — that's intentional and fine as long as no single file imports both. `sqlite.go` imports the parent package to reference `store.Store`/`store.Session`/`store.Turn`/`store.Call`; nothing in the parent package imports `sqlite`. Do not try to merge them into one directory — Go packages are per-directory, so a type's methods must live in the same directory as the type itself (this is why `sqlite/session.go`, `turn.go`, `call.go` all sit flat next to `sqlite.go`, not nested in their own subfolders: they all define methods on `*SQLiteStore`).

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

**`agent`** — Orchestration, split by responsibility across small files:
- `agent.go` — the public surface (`Agent`, `New`, `Handle`) and `run`, `Handle`'s goroutine body: it loads session state, builds the initial request, resolves the mode once, and starts the loop. This is "the turn as seen from outside."
- `state.go` — `State`, the per-request struct threaded through a turn. `ApprovedProposal` here is always in-memory only (see "Mode-driven turns" under [Key patterns](#key-patterns)).
- `input.go` — the `TurnInput` contract (plain message vs. tool result) together with the functions that interpret it into the initial `llm.Request` (`prepareRequest` and its two scenario handlers). Type and the logic that consumes it live in the same file, same pattern as `events.go`.
- `events.go` — the `Event`/`EventType` SSE contract, plus `emit` (send unless ctx is cancelled) and `fail` (emit an error event and report whether one occurred). `fail` is used only by the ReAct loop's own files (`loop.go`, `step.go`, `call.go`); `agent.go` and `input.go` keep their explicit `emit(...)` calls instead of adopting `fail` — a deliberate scoping choice, not an oversight.
- `loop.go` — just the bounded retry skeleton (`runLoop`): call `runStep`, keep going while it says to, stop and report `"max iterations reached"` if it never does.
- `step.go` — `runStep`, one full model-response cycle (stream → consume → persist), plus `consume`/`streamResult` (draining one stream) and `saveAgentTurn` (persisting the agent's turn produced by a step).
- `call.go` — `handleCall`, everything that happens once a step's response includes a tool call: persist it, emit `tool_call`, and either stop (non-auto, client resolves it) or execute-and-re-feed (auto).

Does not know what the modes do internally — it talks to them only through the shared mode interface.

**`modes`** — Business logic. Each mode is a self-contained unit implementing the shared mode interface (`mode.Interface`, in the `mode` subpackage). A mode bundles: instructions (system prompt), tool definitions (what the LLM can call), and tool execution (what happens when the LLM calls a tool). Each mode lives in its own subpackage (`planning/`, `executing/`) with its own files for instructions, tools, and execution. There are exactly two modes and that is fixed for the scope of this project, so there is no registry/lookup-by-ID layer: `agent.resolveMode(approvedProposal)` (in `agent.go`) just constructs `executing.Mode{}` or `planning.Mode{}` directly and returns it as `mode.Interface`. The active mode is **derived, never persisted** — computed once per request. There is no transition/`Next` method: the Planning → Executing move is a one-way boolean flip, resolved pre-loop.

**`llm`** — OpenAI adapter. Wraps the SDK behind a `Client` interface. Translates between agent-level types (`Request`, `StreamEvent`) and SDK types (`ResponseNewParams`, stream events). Sets `ParallelToolCalls: false` (one function call per response). The only package that imports `github.com/openai/openai-go/v3`. `ToolDef.AutoReFeed` is agent metadata and is stripped here — it never reaches OpenAI.

**`store`** — Persistence boundary. `interface.go` (package root) defines the `Store` interface and the domain types (`Session`, `Turn`, `Call`) — nothing here knows SQL exists. `sqlite/` is the (only, for now) implementation, split one file per entity (`session.go`, `turn.go`, `call.go`) plus `sqlite.go` (construction/lifecycle: `SQLiteStore`, `NewSQLite`, `Close`) and `migrations.go` (schema + `RunMigrations`), all in the same package since they all define methods on `*SQLiteStore`. The only layer that imports the SQLite driver. `Session` carries no mode/approval signal — see [Database](#database).

### Key patterns

**Mode-driven turns.** At the start of a request the mode is resolved once from `State.ApprovedProposal` via `agent.resolveMode`, then held fixed for the whole loop. The loop reads `Instructions()` and `Tools()` from it fresh on every iteration. Instructions and tools are NOT inherited between turns — they are explicitly re-sent each time. This prevents hallucination from stale context and enables each mode to fully control what the model sees.

`ApprovedProposal` is computed, not read: `StateFromStore` always starts it at `false` (the store has no such column), and `prepareToolResultRequest` (in `input.go`) flips it to `true` in memory, for the rest of the current request only, when the incoming tool result carries `"approved": true`. Nothing is written back to the store. This means the flip does **not** survive past the request it happened in — the very next request starts back at Planning unless that request's own tool result approves again. This is a deliberate simplification of the scaffold, not a bug: see [Database](#database).

**Turn vs. Step — two different granularities, don't conflate them.** This codebase uses "turn" for two things that are easy to mix up:
- A **`store.Turn`** is one persisted row: one utterance, tagged by `role` (`'user'` or `'agent'`). Standard dialogue-systems usage — every message either side "says" is a turn.
- A **client-visible turn** is one call to `Handle`/`run` — one HTTP request, one round trip from the client's point of view.

These are not 1:1. A single client-visible turn can persist *several* `store.Turn` rows, because the ReAct loop can chain the model into further responses without ever going back to the client: whenever a tool call has `auto_re_feed = true`, `runStep`/`handleCall` execute it, feed the result back to the model automatically, and loop again — each pass is a **Step** (`runStep`, bounded by `maxIterations`), and each Step persists its own `store.Turn` via `saveAgentTurn`. The loop only returns control to the client when the model produces final text, or emits a tool call with `auto_re_feed = false`. So: **Turn = one persisted row. Step = one loop iteration that produces a Turn. One client-visible request can drive many Steps, hence many Turns.**

**`auto_re_feed` drives the loop.** Each tool declares a static `AutoReFeed` bool. When the model emits a call, `handleCall` looks it up (`mode.AutoReFeed`): `true` → execute the tool, record the result, and re-feed automatically (the loop takes another Step); `false` → emit the enriched `tool_call` SSE event (with `call_id` + `auto_re_feed`) and stop the turn so the client resolves it and sends the result back in the next request. Client-sourced tool results and loop-sourced auto results travel the exact same `llm.Request{ToolResults, PrevResponseID}` path — one Request type, two sources.

**One row per call.** A call is INSERTed once (result NULL) when emitted, then UPDATEd in place when it resolves. `PendingCall` finds the single unresolved non-auto call on a session's latest agent turn.

**Streaming via channels.** `agent.Handle()` returns a `<-chan Event`. The loop goroutine writes events as they happen; the HTTP handler reads and translates them to SSE. The agent never writes HTTP directly.

**Context cancellation chain.** `r.Context()` propagates from HTTP client → agent loop goroutine → LLM stream goroutine. All goroutines respect `ctx.Done()` via `select`. When the client disconnects, everything tears down cleanly. No orphaned goroutines.

**Stateless agent.** The `Agent` struct has no per-session state. State is loaded from SQLite at the start of each request (`agent.go`'s `run`) and written back as the request progresses (turns/calls are persisted; `ApprovedProposal` is not — see above). A single `Agent` instance serves all sessions safely.

**Mode interface contract:**

```go
type Interface interface {
    Instructions() string
    Tools() []llm.ToolDef
    Execute(name, args string) (string, error)
}
```

Every mode implements these three methods. The loop only calls these — it never reaches into mode internals. There is deliberately no transition/`Next` method: mode is derived from `ApprovedProposal`, and the Planning → Executing flip is resolved pre-loop, not on the interface.

### Conventions

- **Small files, one responsibility each.** No file should exceed ~200 lines. If it grows, split by sub-responsibility — as plain files in the same package by default (see every file under `agent/` and `store/sqlite/`); reach for a new subpackage only when there's a real polymorphic boundary with multiple concrete implementations (that's why `modes/` has subpackages and `agent/`'s own files don't).
- **Interfaces defined by the provider package**, next to the domain types they gate (`store.Store` + `Session`/`Turn`/`Call` in `store/interface.go`; `llm.Client` in `llm`). Consumers depend on the interface, not the concrete type. The concrete implementation is free to live one directory deeper (`store/sqlite`) without breaking this rule, since Go interface satisfaction is structural, not nominal.
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

- `session` — conversation container bound to a workspace: `session_id`, `root_dir`, `created_at`. That's it — **no mode/approval column**. `ApprovedProposal` is computed in-memory per request in the `agent` layer (see "Mode-driven turns" under [Key patterns](#key-patterns)); it is intentionally never written to `session` or anywhere else in SQLite. A live-coding follow-up that wants the approval to survive across requests would need to add real persistence for it — today it does not, by design.
- `turn` — each user message or agent response. `output_text` holds a user's message or an agent's streamed text; it is NULL for a user turn that carries a tool result (that payload lives in `call.result`). See "Turn vs. Step" under [Key patterns](#key-patterns) for how many rows one client request can produce.
- `call` — one row per tool call. INSERTed with `result` NULL when the agent emits the call, UPDATEd in place when it resolves. Carries `auto_re_feed` so pending client-resolved calls can be found.

There is no persisted `phase`/`mode` column — the mode is always derived from `ApprovedProposal`, and that boolean itself is never persisted either (a stronger statement than earlier versions of this doc made: it used to say `approved_proposal` was "the only persisted mode signal" — it no longer exists in the schema at all). Tables created via `CREATE TABLE IF NOT EXISTS` on startup. No migration framework. Schema lives in `store/sqlite/migrations.go`.

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

There are exactly two modes; the model is *while `ApprovedProposal` is false the agent plans; once it is true the agent executes*. Each mode lives in `internal/agent/modes/<name>/` with three files — `mode.go`, `instructions.go`, `tools.go` — and implements the mode interface (`var _ mode.Interface = Mode{}` for a compile-time check). Since there are exactly two, `agent.resolveMode` in `agent.go` picks the implementation directly with an `if`/`else` on `ApprovedProposal` — no registry, no `ID` type, no lookup-by-name indirection.

The mode boundary is structural: `RunExecution` is present only in Executing's `Tools()` and dispatch; `ProposeStructure` is present only in Planning's. Adding a genuinely new mode would mean a new `ID` constant + `Resolve` branch, but the current design is fixed at two — do not add a third speculatively.

**Scaffold status:** tool handlers are stubs (`"not yet implemented"`) and both instruction strings are empty. The system prompts and tool execution logic are the intended next step, written on top of this scaffold.
