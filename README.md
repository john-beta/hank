# Workspace curator agent

### What does

Reorganize your messy workspace into a cleaner structure.

### Architecture (divide-and-conquer)

_Two modes:_

1. **Planning (human-in-the-loop):** Focus in explore/observe -> propose -> feedback LOOP. We ensure to iterate a solid plan before of procede.
2. **Executing:** Post-planning. Implement the changes. Only focus in materialize the plan (prev mode).

Why: Isolate the concerns of the agent. Reduce hallucinations; improve determinism. Each mode with one-defined task.

_LLM & Prompts:_

* Used OpenAI.
* The prompts, models, and rest of params are stored in OpenAI servers - we reference them via `PromptId` with `Version`.

### Code

* **Golang:** All located in `cmd`.

Optionally, you can locate the client (web) side in `client`
