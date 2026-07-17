package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// maxIterations bounds the loop so a misbehaving model cannot spin forever.
const maxIterations = 10

// runLoop drives the bounded ReAct retry loop for a fixed mode. runStep emits
// every terminal event (done, tool_call, error) itself before reporting
// cont=false, so runLoop's only job past that point is to stop.
func (a *Agent) runLoop(ctx context.Context, state *State, m mode.Mode, req llm.Request, out chan<- Event) {
	for i := 0; i < maxIterations; i++ {
		next, cont := a.runStep(ctx, state, m, req, out)
		if !cont {
			return
		}
		req = next
	}

	a.emit(ctx, out, Event{Type: EventError, Error: "max iterations reached"})
}
