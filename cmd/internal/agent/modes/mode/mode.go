package mode

import "github.com/john-beta/hank/cmd/internal/llm"

// Interface is the contract every mode implements. A mode bundles its system
// prompt, its tool definitions, its tool execution, and its auto-re-feed
// lookup behind these four methods. There is deliberately no Next()/transition
// method: the Planning -> Executing move is a boolean flip resolved before the
// loop (see agent.resolveMode), not a step on this interface.
//
// There are exactly two implementations, planning.Mode and executing.Mode —
// fixed by the scope of this project. Do not add a registry/lookup-by-ID
// layer for a third mode speculatively.
type Interface interface {
	Instructions() string
	Tools() []llm.ToolDef
	Execute(name, args string) (string, error)
	// AutoReFeed reports whether the named tool should be executed and
	// re-fed automatically by the loop (true) or handed back to the client
	// to resolve (false). Each mode implements it by scanning its own
	// Tools() — see planning.Mode.AutoReFeed / executing.Mode.AutoReFeed.
	AutoReFeed(name string) bool
}

// RunExploration is the tool definition shared by both modes: exploring the
// workspace is the same action whether the agent is planning or executing, so
// its schema is defined once here instead of being copy-pasted into
// planning/tools.go and executing/tools.go (where it would risk drifting out
// of sync between the two).
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
