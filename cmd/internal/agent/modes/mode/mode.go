package mode

import "github.com/john-beta/hank/cmd/internal/llm"

// ID identifies an operating mode. The mode is never persisted: it is derived
// per request from the in-memory approval boolean via Resolve. There are
// exactly two.
type ID int

const (
	Planning ID = iota + 1
	Executing
)

// Interface is the contract every mode implements. It is unchanged from the
// original phase contract — a mode bundles its system prompt, its tool
// definitions, and its tool execution behind these three methods. There is
// deliberately no Next()/transition method: the Planning -> Executing move is a
// boolean flip resolved before the loop, not a step on this interface.
type Interface interface {
	Instructions() string
	Tools() []llm.ToolDef
	Execute(name, args string) (string, error)
}

// Resolve maps the single domain boolean to an operating mode: while
// approvedProposal is false the agent plans; once it is true the agent
// executes. Nothing here is persisted — the caller computes the boolean fresh
// for each request.
//
// The flip is one-way for now — there is no Executing -> Planning transition;
// that is future work.
func Resolve(approvedProposal bool) ID {
	if approvedProposal {
		return Executing
	}
	return Planning
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
