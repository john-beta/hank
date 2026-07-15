package planning

import (
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// Mode is the Planning mode: active while session.approved_proposal is false. It
// is a stateless value type.
type Mode struct{}

var _ mode.Interface = Mode{}

func (Mode) Instructions() string                      { return instructions }
func (Mode) Tools() []llm.ToolDef                      { return tools }
func (Mode) Execute(name, args string) (string, error) { return execute(name, args) }
