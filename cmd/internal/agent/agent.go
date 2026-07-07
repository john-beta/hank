package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// Agent orchestrates the ReAct loop over an LLM client. It depends only on the
// llm.Client and store.Store interfaces, never on concrete SDK types.
type Agent struct {
	llm   llm.Client
	store store.Store // injected for future use; unused in this scaffold
	state *State
	tools *ToolRegistry
}

// New wires an agent with its LLM client and store.
// TODO: Refactor State: should be scoped by a sessionID using Store, not a single instance per Agent. The agent MUST be stateless, reusable across sessions.
func New(llmClient llm.Client, st store.Store) *Agent {
	return &Agent{
		llm:   llmClient,
		store: st,
		state: NewState(),
		tools: NewToolRegistry(),
	}
}

// Handle starts a turn for userText and returns a channel of events. The
// channel is closed when the turn completes or ctx is cancelled.
func (a *Agent) Handle(ctx context.Context, userText string) <-chan Event {
	out := make(chan Event)

	go a.runLoop(ctx, userText, out)
	return out
}
