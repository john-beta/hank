package mode

import "github.com/john-beta/hank/cmd/internal/llm"

// Interface is the contract every mode implements. A mode bundles its system
// prompt, its tool definitions, and its tool execution behind these three
// methods. There is deliberately no Next()/transition method: the Planning ->
// Executing move is a boolean flip resolved before the loop (see
// agent.resolveMode), not a step on this interface.
//
// There are exactly two implementations, planning.Mode and executing.Mode —
// fixed by the scope of this project. Do not add a registry/lookup-by-ID
// layer for a third mode speculatively.
type Interface interface {
	Instructions() string
	Tools() []llm.ToolDef
	Execute(name, args string) (string, error)
}

// AutoReFeed reports a tool's auto_re_feed flag by looking it up by name in a
// mode's tool set. It is the loop's single source of truth for whether a call
// is executed-and-re-fed automatically or handed back to the client. An unknown
// name reports false.
func AutoReFeed(tools []llm.ToolDef, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return t.AutoReFeed
		}
	}
	return false
}
