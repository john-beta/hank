Tools Definition:

Tool Name: ProposeStructure

Parameters:

- workspace_entries:{
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

ONLY use this tool after the exploration is complete.

Usage:

- Use `workspace_entries` to structure your proposal.

- Include one entry for each file in the proposal.

- Every proposal must contain the complete `workspace_entries` array. If the user requests changes, regenerate and resubmit the entire proposal with the requested updates applied.

Example:

```json
{
  "workspace_entries": [
    { "path": "README.md" },
    { "path": "src/components/Button.vue" },
    { "path": "src/components/Card.vue" },
    { "path": "src/pages/index.ts" }
  ]
}
```
