package planning

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
)

// tools is Planning's tool set: RunExploration (auto-re-fed) and
// ProposeStructure (not auto-re-fed — the client resolves it, and an approving
// result is what flips the session into Executing). RunExecution is
// deliberately absent here; its absence is the mode boundary.
var tools = []llm.ToolDef{
	{
		Name:        "RunExploration",
		Description: "Explore the workspace (not yet implemented).",
		AutoReFeed:  true,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{"type": "string"},
			},
			"required": []string{"code"},
		},
	},
	{
		Name:        "ProposeStructure",
		Description: "Propose a workspace structure (not yet implemented).",
		AutoReFeed:  false,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"structure": map[string]any{
					"type":  "array",
					"items": map[string]any{"$ref": "#/$defs/node"},
				},
			},
			"required": []string{"structure"},
			"$defs": map[string]any{
				"node": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"file": map[string]any{"type": "string"},
						"children": map[string]any{
							"type":  "array",
							"items": map[string]any{"$ref": "#/$defs/node"},
						},
					},
					"required": []string{"file"},
				},
			},
		},
	},
}

// execute routes Planning's known tool names. The routing seam is wired; each
// handler body is a stub returning a placeholder until the business logic is
// written. An unknown name for this mode is an error.
func execute(name, args string) (string, error) {
	switch name {
	case "RunExploration":
		return "not yet implemented", nil
	case "ProposeStructure":
		return "not yet implemented", nil
	default:
		return "", fmt.Errorf("planning: unknown tool: %s", name)
	}
}
