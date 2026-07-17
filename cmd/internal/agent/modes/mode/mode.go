package mode

import "github.com/john-beta/hank/cmd/internal/llm"

// Mode is a plain struct of values, not an interface satisfied by an empty
// struct — with no behavior beyond these four fields, there's nothing for
// method-forwarding boilerplate to buy. planning.New() / executing.New() each
// build one; agent.resolveMode picks between them with a plain if/else — no
// registry, since there are exactly two by design.
type Mode struct {
	Instructions string
	Tools        []llm.ToolDef
	Execute      func(name, args string) (string, error)
	// AutoReFeed reports whether the named tool executes-and-re-feeds
	// automatically. Each mode builds it by closing over its own Tools().
	AutoReFeed func(name string) bool
}

// RunExploration is shared by both modes — exploring the workspace behaves
// the same regardless of mode — so its schema lives here once instead of
// risking drift between planning/tools.go and executing/tools.go.
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
