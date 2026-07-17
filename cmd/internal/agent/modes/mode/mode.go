package mode

import "github.com/john-beta/hank/cmd/internal/llm"

// Mode bundles a mode's instructions, tools, and execution behavior.
// planning.New() / executing.New() each build one.
type Mode struct {
	Instructions string
	Tools        []llm.ToolDef
	Execute      func(name, args string) (string, error)
	// AutoReFeed reports whether the named tool executes-and-re-feeds
	// automatically. Each mode builds it by closing over its own Tools.
	AutoReFeed func(name string) bool
}

// RunExploration is shared by both modes, so its schema lives here once.
var RunExploration = llm.ToolDef{
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
}
