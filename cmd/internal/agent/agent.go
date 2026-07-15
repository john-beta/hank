package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// Agent orchestrates the ReAct loop over an LLM client. It is stateless: all
// conversation state lives in the store and is loaded per request, so a single
// Agent is safely reused across sessions. It depends only on the llm.Client and
// store.Store interfaces, never on concrete SDK types.
type Agent struct {
	llm   llm.Client
	store store.Store
}

// New wires an agent with its LLM client and store.
func New(llmClient llm.Client, st store.Store) *Agent {
	return &Agent{
		llm:   llmClient,
		store: st,
	}
}

// Handle starts a turn for the given input within sessionID and returns a
// channel of events. The channel is closed when the turn completes or ctx is
// cancelled.
func (a *Agent) Handle(ctx context.Context, sessionID string, input TurnInput) <-chan Event {
	out := make(chan Event)
	go a.run(ctx, sessionID, input, out)
	return out
}
