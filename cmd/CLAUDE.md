# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Scope: this document covers the **server-side Go application under `cmd/`**. The Vue client under `client/` is out of scope here.

## Commands

All Go commands run from the `cmd/` directory (that's where `go.mod` lives).

```sh
cd cmd
go build ./...              # build
go vet ./...                # vet
go run ./server            # run the server (listens on :8080)
```

Runtime requirements:
- `OPENAI_API_KEY` must be set — `main.go` calls `log.Fatal` without it.
- **Windows only.** Tool execution isolates child processes with a Windows Job Object (`go-winjob`) and shells out to the `py` launcher, so the server does not build/run on non-Windows hosts.
- On start it opens/creates `hank.db` (SQLite) in the working directory and applies the schema idempotently.

There are currently **no tests** in the Go codebase.

## Architecture

Hank is an agentic coding assistant. The server drives a two-phase LLM agent (plan, then execute) over the OpenAI Responses API, persisting every turn and tool call to SQLite, and streams events to the client as NDJSON.

### Layers (dependency direction points inward)

```
transport/http  →  agent  →  llm      (OpenAI boundary)
                     │     →  store    (SQLite persistence)
                     └────  agent/modes (per-phase prompt + tool routing)
```

- **`transport/http`** — routing, CORS (only `http://localhost:5173`, the Vite dev origin), request DTOs, and NDJSON streaming. Handlers are thin: they decode, call `agent.Handle`, and forward each `agent.Event` as one JSON line, flushing per frame. The request `context` is the cancellation signal — a client disconnect unwinds the agent loop and the upstream LLM stream.
- **`agent`** — the ReAct orchestration loop (the core of the system; see below).
- **`agent/modes`** — the two agent behaviors. Each `Mode` bundles an OpenAI prompt reference with Go-side tool routing (`Execute`) and per-tool `AutoReFeed` policy.
- **`llm`** — a narrow `Client` interface (`Stream`) wrapping the OpenAI SDK. **No OpenAI SDK types cross this boundary** — `Request`/`StreamEvent`/`Event` are all local types, so the agent never depends on the SDK directly.
- **`store`** — a `Store` interface with a single pure-Go SQLite implementation (`modernc.org/sqlite`, no cgo).

### The agent loop (`agent/loop.go`, `agent/agent.go`)

`Agent` is **stateless** — one instance serves all sessions. All conversation state lives in the store and is rebuilt fresh per request (`StateFromStore`). `Handle` spawns a goroutine that emits `Event`s on a channel which closes when the turn ends or `ctx` is cancelled.

Each turn: resolve state → `prepareRequest` → resolve mode → run a **bounded ReAct loop** (`maxIterations = 15`). Each step streams one model response and either:
1. ends the turn (no tool call → save assistant turn, emit `done`), or
2. handles a tool call via `handleCall`.

Key invariants:
- `ParallelToolCalls` is **off**, so a step yields at most one function call.
- The stored prompt is **re-set on every step** (`req.Prompt = m.Prompt`) because OpenAI's `PreviousResponseID` does not carry it forward.
- Response chaining uses `PrevResponseID`, not a re-sent message history.

### Planning → Executing: the mode flip

There are exactly two modes and the transition is **one-way** (no Executing → Planning). The mode is *computed per request*, never persisted — `resolveMode(state.ApprovedProposal)`.

- **Planning** (`modes/planning.go`): tools `RunExploration` (auto-re-fed) and `ProposeStructure` (**not** auto-re-fed — client-resolved). `RunImplementation` is deliberately absent; its absence *is* the mode boundary.
- **Executing** (`modes/executing.go`): tools `GetWorkspaceCurrentState` and `RunImplementation`, both auto-re-fed. Its presence of `RunImplementation` is the other half of the boundary.

The flip happens in `prepareToolResultRequest` (`agent/input.go`): when the client submits a `ProposeStructure` result whose JSON has `{"approved": true}`, the request sets `ApprovedProposal = true`, **clears `PrevResponseID`** (so Executing starts fresh rather than being biased by planning-phase reasoning), and sends a plain `"Plan approved."` message. On that same request, `resolveModeVariables` loads the proposal's `proposed_workspace_entries` from the stored call args and injects it into the Executing prompt as the `plan_structure` variable.

### auto_re_feed: the human-in-the-loop mechanism

Every tool call is classified `AutoReFeed` true/false per the active mode's policy:
- **Auto-re-fed** tools execute server-side immediately (`m.Execute`), the result is persisted and emitted, and the loop continues by feeding the output back to the model.
- **Non-auto** tools (only `ProposeStructure`) stop the loop after emitting `tool_call`. The client resolves them and submits the result on a *new* request. This is the approval gate.

`AutoReFeed` is emitted as `*bool` so a meaningful `false` survives JSON (a plain `false` would be dropped by `omitempty`), telling the client it must resolve the call.

### Tool execution (`modes/child_process.go`, `modes/job_object.go`)

All four tools ultimately run a Python script (passed as the `script` arg) via `py -I -` (isolated mode, script on stdin) with `cmd.Dir` set to the session's `RootDir`. The child runs inside a Windows **Job Object** (`runInJobObject`) that enforces: kill-on-close, an active-process limit of 2, and a 256 MiB memory cap, with a 15s `WaitDelay`. The environment is stripped to just `SystemRoot`. Results are returned as a JSON `PythonResult` (`success`/`output`/`error`) — a failing tool returns a JSON error string, not a Go error, so the model can react to it.

### Persistence model (`store/sqlite/migrations.go`)

Three tables, created idempotently at startup (no migration framework):
- **`session`** — a conversation bound to a `root_dir` workspace.
- **`turn`** — one exchange. `output_text` is the user message or assistant text; it is **NULL for a user turn that carries a tool result** (that payload lives in `call.result`). `response_id` is set only for assistant turns.
- **`call`** — one row per `call_id`. Inserted with `result = NULL` when the call is emitted, then **UPDATE-in-place** when it resolves. Never a second INSERT for the same `call_id`.

Derived state (never stored as columns):
- **Session "finished"** = the session has an assistant turn with a `RunImplementation` call (`IsSessionImplemented`). A finished session rejects further turns.
- **Pending call** = a `call` row with `result IS NULL AND auto_re_feed = 0` (`IsPendingCall`). Submitting a tool result for anything else is treated as desync and rejected before it reaches OpenAI.

Note: `LastAssistantTurn` orders by `created_at DESC, rowid DESC` because timestamps have one-second resolution and a single loop can persist several assistant turns within the same second.

## Conventions

- **OpenAI Responses API, not Chat Completions.** Prompts (system instructions, tool schemas, model, reasoning settings) are **stored server-side in the OpenAI dashboard** and referenced only by `ID` + `Version` constants in the `modes` package. The `.md` files at the root of `cmd/` are human-readable references for those dashboard prompts/tools; changing tool behavior often means editing the dashboard prompt, not just Go code.
- Errors surface to the client as an `error` event on the stream, not an HTTP status (once streaming has begun).
- Both the LLM stream goroutine and the agent goroutine send on channels via a `select { case ch <- ev: case <-ctx.Done(): }` guard so they never block on an abandoned consumer.
