package phases

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/agent/phases/phase"
	"github.com/john-beta/hank/cmd/internal/agent/phases/phase_one"
	"github.com/john-beta/hank/cmd/internal/agent/phases/phase_two"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// Get returns the phase implementation for id. An unknown or unset id yields
// emptyPhase, which drives the loop to a clean, tool-less, instruction-less
// completion instead of panicking.
func Get(id phase.ID) phase.Interface {
	switch id {
	case phase.PhaseOneID:
		return phase_one.Phase{}
	case phase.PhaseTwoID:
		return phase_two.Phase{}
	default:
		return emptyPhase{}
	}
}

// emptyPhase is the null object for an unknown/unset phase: no instructions,
// no tools, and any tool call is an error.
type emptyPhase struct{}

var _ phase.Interface = emptyPhase{}

func (emptyPhase) Instructions() string { return "" }
func (emptyPhase) Tools() []llm.ToolDef { return nil }

func (emptyPhase) Execute(name, args string) (string, error) {
	return "", fmt.Errorf("phases: empty phase cannot execute tool %q", name)
}
