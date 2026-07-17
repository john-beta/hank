package planning

import (
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

func New() mode.Mode {
	return mode.Mode{
		Prompt:     llm.Prompt{ID: promptID, Version: promptVersion},
		Execute:    execute,
		AutoReFeed: func(name string) bool { return autoReFeed[name] },
	}
}
