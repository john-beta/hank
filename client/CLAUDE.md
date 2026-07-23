# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Scope: this document covers the **Vue client under `client/`**. The Go server under `cmd/` is out of scope here (it has its own CLAUDE.md).

## Vendored libraries — not our code

Three directories hold **third-party source code vendored into the repo**, not our application logic. Treat them as an external dependency: import from them, but do not document, refactor, or reason about our patterns from them.

- `src/components/ui/` — **shadcn/ui** components (shadcn-vue, "new-york" style). Config in `components.json`.
- `src/components/ai-elements/` — **AI Elements (Vue)** components, which build on shadcn/ui. This is the primary UI library for chat/tool rendering (`message/`, `prompt-input/`, `tool/`, `code-block/`, `file-tree/`).
- `src/lib/` — library utilities (e.g. `cn`), imported as `@/lib/utils`.

These are added via the shadcn/ui CLI, so re-generating a component overwrites local edits. **All of our own logic lives outside these three directories.** Everything below describes only our code.

## Commands

Run from `client/`:

```sh
npm run dev          # Vite dev server (default :5173 — the origin the server's CORS allows)
npm run build        # type-check + production build
npm run type-check   # vue-tsc
npm run format       # prettier over src/
```

`VITE_API_BASE_URL` (in `.env`) points at the Go server (`http://localhost:8080`).

Type-check note: a pre-existing TS5101 `baseUrl` deprecation (TypeScript 6) can abort `vue-tsc`; run `npx vue-tsc --noEmit --ignoreDeprecations 6.0` to check app code in isolation.

## Architecture

A Vue 3 SPA (`<script setup>` + TS, Vite, Tailwind v4) that is a thin chat front-end over the server's streaming agent. The design centers on turning the server's flat NDJSON event stream into a turn-centric UI model, and on a human-in-the-loop approval gate for certain tool calls.

### Three layers, one direction

```
services/      wire protocol — fetch, NDJSON parsing, event→model folding
   ↑
composables/   application state (singletons) — session, message stream, approval
   ↑
components/    rendering only — read composable state, emit user intent
```

- **`services/`** — `api.ts` (create session), `streaming.ts` (POST `/api/messages`, read the NDJSON body line-by-line into `StreamEvent`s), `aggregateStream.ts` (the fold — see below).
- **`composables/`** — the app's state stores. See the singleton pattern below.
- **`components/`** — presentational. `App.vue` → `MainPanel.vue` switches between the two top-level states.

### Singleton composables (important)

`useChatSession`, `useMessageStream`, and `useToolApproval` all declare their state with **module-level `ref`s**, so every component that calls the composable shares **one** instance — they act as global stores, not per-component state. This is why, e.g., `ChatFooter` and the `HITL` tool part can both drive `resolveTool`/approval without prop-drilling. Keep this in mind: mutating state in one component is observed everywhere.

### Two top-level UI states

`MainPanel.vue` renders by session presence (`useChatSession.isWizardMode`):
- **Wizard** (no session): `ChatWizard.vue` creates a session from a chosen `root_dir`.
- **Chat** (active session): `ChatHeader` + `ChatBody` + `ChatFooter`.

### Event stream → UI model (the fold)

The server streams a flat sequence of events (`text_delta`, `tool_call`, `tool_result`, `error`, `done`). `aggregateStream.ts#applyEvent` folds them into `ChatMessage[]`, where one assistant response becomes **one message with ordered `parts`** (text, tool, error). It appends deltas onto the trailing text part, and matches `tool_result` back to its `tool_call` part by `call_id`. The UI types (`ChatMessage`/`Part`) live in `types.ts` alongside the wire types (`StreamEvent`/`TurnInput`) — keep the two clearly separated when editing.

### Part rendering is a registry

`MessageRenderer` → `PartRenderer` maps `part.type` to a component via `COMPONENT_BY_TYPE`. **To add a renderable part type:** (1) add the variant to the `Part` union in `types.ts`, (2) create the component in `Message/parts/`, (3) register it in `PartRenderer`'s map. `ToolPart` then sub-routes on `part.autoReFeed`:
- `true` → `tool/AutoRefeed.vue` — a server-executed tool; renders the `script` arg as a Python `CodeBlock` and the result (`{success, output|error}`) as JSON, or the error text in a raw `<pre>`.
- `false` → `tool/NoAutoRefeed.vue` — the `ProposeStructure` plan; builds a nested `FileTree` from the flat `proposed_workspace_entries` paths, and embeds the `HITL` approval control.

### Human-in-the-loop approval (the gate)

A `tool_call` with `auto_re_feed === false` stops the agent loop server-side until the user resolves it. The client mirrors that:

1. `useMessageStream.run` sees the non-auto `tool_call` and calls `approval.start(sessionId, callId)`.
2. `useToolApproval` marks a pending call; `inputBlocked` locks `ChatFooter` (placeholder changes to "Approve or request changes…").
3. In the pending tool's `HITL` control the user either **Approve**s (`resolveTool(true, 'Plan approved')`) or clicks **Request changes** (`unlock()`), which unlocks the footer so their next submit becomes the rejection message.
4. `ChatFooter.handleSubmit` checks `pending`: if set, the text resolves the call via `resolveTool(false, text)`; otherwise it's a fresh `sendMessage`.

Both paths funnel through `useMessageStream.run(sessionId, echo, input)` — a plain message and a tool decision differ only in the `TurnInput` body sent and the text echoed as the user bubble. The tool-result body is `{ tool_result: { call_id, result: { approved, message } } }`; the server flips into its Executing phase when `approved` is true.

### Conventions

- Import our own code by relative path; import vendored libraries and `@/lib/utils` via the `@/` alias (see `vite.config.ts` / `components.json` aliases).
- `ChatBody` auto-scrolls only when the user is already near the bottom (`SCROLL_BOTTOM_OFFSET`), so reading back isn't interrupted; it resets the stream on `sessionId` change.
- Tool wire shapes are stable and relied upon: auto-re-feed args are `{ "script": "<python>" }`; results are `{ success, output }` or `{ success, error }`; `ProposeStructure` args are `{ "proposed_workspace_entries": [{ "path" }] }`.
