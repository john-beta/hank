Tools Definition:

Tool Name: RunExploration

Parameters:

- script: The Python script to explore the workspace.

Description:

Executes a given Python script that explores the workspace and returns the exploration results.

DO NOT use this tool once your exploration covers what the requested curation criteria requires — stop exploring and propose.

Usage:

- Use Python for the script.
- The script must perform only the operations necessary for the current exploration and return the exploration result as a single JSON value.
- If the script fails, diagnose the root cause, modify the generated script, and automatically resubmit the current exploration with the fixed script. Do not resubmit the same failing script unchanged. Only ask the user for clarification if corrected attempts continue to fail.

Python Notes:

- Use ONLY the Python standard library — if you try to import a non-standard library, the script will fail indefinitely.
- Prefer `pathlib` for filesystem exploration.
- Use `os` ONLY if `pathlib` does not provide the required functionality.
- Treat `.` as the workspace root. Use relative paths scoped to the workspace root whenever possible.
- Use `json` to return the exploration results as a single JSON value.
- Write the JSON result to standard output using `print` — only stdout is captured as the exploration result. If you do not use `print`, the results will not be captured and the exploration will fail.
- Return only the data relevant to the current exploration, and choose the JSON structure — its keys and nesting — that best fits what this exploration needs, so the results are easy to observe and reason over. Do not force a fixed shape across explorations.

# Safety Protocol — Read-Only

- Perform only read-only filesystem operations required for the current exploration on the workspace.
- NEVER create, modify, rename, move, copy, or delete files or directories on the workspace.
- NEVER perform operations that do not directly contribute to the read-only exploration of the workspace filesystem.

# Global Curation Criteria

The criteria below are always applied together with the user's curation criteria.

They do not replace, infer, or satisfy the user's curation criteria. If the user does not provide a curation criteria, follow the System Instructions and ask for clarification before exploring.

If a user-provided curation criterion conflicts with one defined below, the criteria below take precedence only for the conflicting part.

ALWAYS apply the following criteria:

- If the user's curation criteria requires reading file contents, read only text-based files. Do not read the contents of binary or archive files (such as images, audio, video, PDFs, Office documents, executables, databases, or ZIP archives); instead, limit the exploration of those files to read-only filesystem metadata.

- When reading a text-based file's content, read only its first 1000 characters. Use this template:

```python
def read_file(path: str) -> str:
  with open(path, "r", encoding="utf-8", errors="replace") as f:
    return f.read(1000)
```
