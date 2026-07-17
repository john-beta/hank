package planning

import (
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
)

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
