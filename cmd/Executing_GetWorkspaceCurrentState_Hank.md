Tools Definition:

Tool Name: GetWorkspaceCurrentState

Parameters:

- script: The Python script to get the current state of the workspace.

Description:

Executes a given Python script that returns the current state of the workspace.

Usage:

- Use Python for the script.
- The script must perform only the operations necessary to obtain the current state of the workspace and return it as a single JSON value.
- If the script fails, diagnose the root cause, modify the generated script, and automatically resubmit with the fixed script.

Python Notes:

- Use ONLY the Python standard library — if you try to import a non-standard library, the script will fail indefinitely.
- Prefer `pathlib` for filesystem operations.
- Use `os` ONLY if `pathlib` does not provide the required functionality.
- Treat `.` as the workspace root. Use relative paths scoped to the workspace root whenever possible.
- Use `json` to return the current state of the workspace as a single JSON value.
- Write the JSON result to standard output using `print` — only stdout is captured as the current state of the workspace.
- Return the current state of the workspace as a JSON object containing every workspace entry, each represented by its path.

# Safety Protocol — Read-Only

- Perform only read-only filesystem operations required to obtain the current state of the workspace.

- NEVER perform mutations on the workspace.

- NEVER perform operations that do not directly contribute to obtaining the current state of the workspace.
