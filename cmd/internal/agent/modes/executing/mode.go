package executing

import (
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
)

// New builds the Executing mode: active once the approval boolean is true.
func New() mode.Mode {
	return mode.Mode{
		Instructions: instructions,
		Tools:        tools,
		Execute:      execute,
		AutoReFeed: func(name string) bool {
			for _, t := range tools {
				if t.Name == name {
					return t.AutoReFeed
				}
			}
			return false
		},
	}
}
