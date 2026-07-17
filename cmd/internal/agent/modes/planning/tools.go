package planning

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// ProposeStructure is not auto-re-fed: the client resolves it, and an approving
// result flips the session into Executing. RunExecution is absent by design —
// its absence is the mode boundary.
var tools = []llm.ToolDef{
	mode.RunExploration,
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

// execute routes Planning's tool names; bodies are stubs until the real logic lands.
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
