Tools Definition:

Tool Name: ProposeStructure

Parameters:

- proposed_workspace_entries:{
  description:'Proposed workspace structure after exploration, organized according to the requested curation criteria.',
  type:'array',
  items: {
  type:'object',
  properties: {
  path: {
  description: 'Relative path to the file from the workspace root. Always include the complete directory hierarchy ending with the file name (e.g. "folder1/folder2/file.txt").',
  type: 'string',
  },
  },
  required: ['path'],
  },
  }

Description:

Proposes a new workspace structure based on the exploration results.

Use this tool only after exploration is complete, to propose a new structure — not to describe the current workspace.

Usage:

- Use `proposed_workspace_entries` to structure your proposal.

- Include one entry for each file in the proposal.

- Every entry must be a file path; folders exist only as segments within a file's path, never as standalone entries. Do not propose an empty folder — a folder appears in the structure only when it contains at least one file. If the curation criteria would require a folder with no files, do not invent a placeholder file: explain the situation to the user instead.

- Every proposal must contain the complete `proposed_workspace_entries` array. If the user requests changes, regenerate and resubmit the entire proposal with the requested updates applied.

- Since proposal approvals are only valid via this tool, if the user approves a proposal outside of it, you MUST regenerate the last proposal and resubmit it via the tool. Include an explanatory text asking to confirm it officially.

# Safety Protocol

- Every entry MUST have an unique path.
- If the curation criteria would produce duplicates, explain the conflict to the user with alternatives. Do not propose until the duplication conflict is resolved.

# Example

```json
{
  "proposed_workspace_entries": [
    { "path": "README.md" },
    { "path": "src/components/Button.vue" },
    { "path": "src/components/Card.vue" },
    { "path": "src/pages/index.ts" }
  ]
}
```
