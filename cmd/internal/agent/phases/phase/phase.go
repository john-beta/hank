package phase

import "github.com/john-beta/hank/cmd/internal/llm"

// ID identifies a phase. It is backed by int so it maps directly to the
// persisted session.phase column.
type ID int

// Interface is the contract every phase implements. It is the extension point
// for Task 2 business logic: a new phase is a new implementation of this
// interface, requiring no change to the ReAct loop.
type Interface interface {
	Instructions() string
	Tools() []llm.ToolDef
	Execute(name, args string) (string, error)
}

const (
	PhaseOneID ID = iota + 1
	PhaseTwoID
	// Task 2 adds more here.
)
