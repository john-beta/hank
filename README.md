# Workspace curator agent

### What does

Reorganize your messy workspace into a cleaner structure.

### Design (divide-and-conquer)

_Two modes:_

1. **Planning (human-in-the-loop):** Focus in explore/observe -> propose -> feedback LOOP. We ensure to iterate a solid plan before of procede.
2. **Executing:** Post-planning. Implement the changes. Only focus in materialize the plan (prev mode).

**Code as action:** both modes act by writing Python that the server runs in a sandboxed child process. Exploring, proposing, and applying changes are executed code, not fixed tool calls.

### Stack

- **Go** server in `cmd` (the agent). **Vue** client in `client`.
- LLM is OpenAI. Prompts, models, and params live on OpenAI's servers, referenced by `PromptId` + `Version`.

### Run

Server:

```sh
cd cmd\server && go run .
```

Client:

```sh
cd client && npm run dev
```

Requires the `OPENAI_API_KEY` environment variable. The account behind that key must also host the agent's prompts — replicate them following [prompts/README.md](cmd/internal/agent/modes/prompts/README.md).

> Prompts are still evolving. The files under `prompts/` are kept in sync as reference whenever they change.
