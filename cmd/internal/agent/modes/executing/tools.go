package executing

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
)

// tools is Executing's tool set: RunExploration (auto-re-fed) and RunExecution
// (auto-re-fed). RunExecution is kept a distinct tool from RunExploration — its
// presence here and absence from Planning is the mode boundary, even though the
// two may later share an executor.
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

// execute routes Executing's known tool names. The routing seam is wired; each
// handler body is a stub returning a placeholder until the business logic is
// written. An unknown name for this mode is an error.
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
