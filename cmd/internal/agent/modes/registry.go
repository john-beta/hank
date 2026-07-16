package modes

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/agent/modes/executing"
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/agent/modes/planning"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// Get returns the mode implementation for id. An unknown id yields emptyMode,
// which drives the loop to a clean, tool-less, instruction-less completion
// instead of panicking. The id itself is produced by mode.Resolve from the
// in-memory approval boolean.
func Get(id mode.ID) mode.Interface {
	switch id {
	case mode.Planning:
		return planning.Mode{}
	case mode.Executing:
		return executing.Mode{}
	default:
		return emptyMode{}
	}
}

// emptyMode is the null object for an unknown mode: no instructions, no tools,
// and any tool call is an error.
type emptyMode struct{}

var _ mode.Interface = emptyMode{}

func (emptyMode) Instructions() string { return "" }
func (emptyMode) Tools() []llm.ToolDef { return nil }

func (emptyMode) Execute(name, args string) (string, error) {
	return "", fmt.Errorf("modes: empty mode cannot execute tool %q", name)
}
