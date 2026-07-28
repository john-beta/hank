# Prompts

The `planning/` and `executing/` folders mirror the prompts this agent expects in OpenAI. Each folder holds one system instruction (`.md`) and its tools (`.json`).

The agent references prompts by ID, so you must recreate them under your own OpenAI account.

## Replicate the prompts

Do this once per folder (`planning/` and `executing/`):

1. Open the [OpenAI Playground](https://platform.openai.com/chat/edit).
2. Paste the `.md` content into the system instructions.
3. For each `.json` file, add a tool under **Local tools > Add Function**.
4. **`executing/` only:** under **Variables > Add**, add a variable named exactly `proposed_structure`.
5. Save the prompt and copy its `prompt_id` and `version`.

## Wire them up

Set the IDs and versions you copied:

- `planning/` → `planningPromptID` and `planningPromptVersion` in `planning.go`.
- `executing/` → `executingPromptID` and `executingPromptVersion` in `executing.go`.

Then run the agent.
