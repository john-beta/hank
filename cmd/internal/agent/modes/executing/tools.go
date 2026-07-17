package executing

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// RunExecution is kept a distinct tool from RunExploration — its presence
// here and absence from Planning is the mode boundary, even though the two
// may later share an executor.
var tools = []llm.ToolDef{
	mode.RunExploration,
	{
		Name:        "RunExecution",
		Description: "Execute the approved restructure (not yet implemented).",
		AutoReFeed:  true,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{"type": "string"},
			},
			"required": []string{"code"},
		},
	},
}

// execute routes Executing's known tool names. Handler bodies are stubs until
// the real logic is written; an unknown name is an error.
func execute(name, args string) (string, error) {
	switch name {
	case "RunExploration":
		return "not yet implemented", nil
	case "RunExecution":
		return "not yet implemented", nil
	default:
		return "", fmt.Errorf("executing: unknown tool: %s", name)
	}
}
