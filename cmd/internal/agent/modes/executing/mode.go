package executing

import (
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// Mode is the Executing mode: active once the approval boolean is true. It
// is a stateless value type.
type Mode struct{}

var _ mode.Interface = Mode{}

func (Mode) Instructions() string                      { return instructions }
func (Mode) Tools() []llm.ToolDef                      { return tools }
func (Mode) Execute(name, args string) (string, error) { return execute(name, args) }

// AutoReFeed reports whether the named tool auto-re-feeds, by scanning this
// mode's own tool set.
func (Mode) AutoReFeed(name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return t.AutoReFeed
		}
	}
	return false
}
