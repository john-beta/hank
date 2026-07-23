Tools Definition:

Tool Name: RunImplementation

Parameters:

- script: The Python script to implement the changes over the workspace.

Description:

Executes a given Python script that implements the changes over the workspace and returns the implementation results.

Usage:

- Use Python for the script.
- The script must perform only the operations necessary for the implementation and return the implementation result as a single JSON value.
- If the script fails due to an execution error and cannot produce a valid JSON result, diagnose the root cause, modify the generated script, and automatically resubmit the implementation with the fixed script. If corrected attempts remain to fail, stop the implementation and report the error to the user.

Python Notes:

- Use ONLY the Python standard library — if you try to import a non-standard library, the script will fail indefinitely.
- Prefer `pathlib` for filesystem operations.
- Use `os` ONLY if `pathlib` does not provide the required functionality.
- Treat `.` as the workspace root. Use relative paths scoped to the workspace root whenever possible.
- Use `json` to return the implementation results as a single JSON value.
- Write the JSON result to standard output using `print` — only stdout is captured as the implementation result.
- Return a JSON object that audits every individual change operation in this implementation. Wrap each filesystem action in its own `try/catch` and record, for each operation, the target, the action, and `success: true | false` plus any error message. Do not add recovery logic here.

# Safety Protocol

- Perform ONLY the mutations specified exactly in the plan structure. Additional mutations are not allowed.

- Perform ONLY the filesystem operations required for the implementation.
