package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/modes/executing"
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/agent/modes/planning"
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

// run is Handle's goroutine body: load session state, build the initial LLM
// request from input (this is where the possible Planning -> Executing
// transition happens, before the loop, never inside it), resolve the mode
// once, then run the loop. It owns the output channel and always closes it.
func (a *Agent) run(ctx context.Context, sessionID string, input TurnInput, out chan<- Event) {
	defer close(out)

	sess, err := a.store.GetSession(ctx, sessionID)
	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	lastAgentTurn, err := a.store.LastAgentTurn(ctx, sessionID)
	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	state := StateFromStore(sess, lastAgentTurn)

	req, ok := a.prepareRequest(ctx, state, input, out)
	if !ok {
		return // an error event was already emitted
	}

	// Mode is derived from the (possibly just-flipped) approval boolean, resolved
	// once and held fixed for the whole loop.
	m := resolveMode(state.ApprovedProposal)

	a.runLoop(ctx, state, m, req, out)
}

// resolveMode maps the single domain boolean to a mode implementation: while
// approvedProposal is false the agent plans; once it is true the agent
// executes. There are exactly two modes by design (see mode.Interface) — no
// registry/lookup-by-ID layer needed for that.
//
// The flip is one-way for now — there is no Executing -> Planning transition;
// that is future work.
func resolveMode(approvedProposal bool) mode.Interface {
	if approvedProposal {
		return executing.Mode{}
	}
	return planning.Mode{}
}
