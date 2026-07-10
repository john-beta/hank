package phase_two

import (
	"github.com/john-beta/hank/cmd/internal/agent/phases/phase"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// Phase is the tool-less second phase. It is a stateless value type.
type Phase struct{}

var _ phase.Interface = Phase{}

func (Phase) Instructions() string                      { return instructions }
func (Phase) Tools() []llm.ToolDef                      { return tools }
func (Phase) Execute(name, args string) (string, error) { return execute(name, args) }
