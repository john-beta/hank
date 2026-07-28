package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/modes"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// Agent is stateless: all state lives in the store, so one instance serves all
// sessions.
type Agent struct {
	llm   llm.Client
	store store.Store
}

func New(llmClient llm.Client, st store.Store) *Agent {
	return &Agent{
		llm:   llmClient,
		store: st,
	}
}

// Handle returns a channel of events; it closes when the turn completes or ctx
// is cancelled.
func (a *Agent) Handle(ctx context.Context, sessionID string, input TurnInput) <-chan Event {
	out := make(chan Event)
	go a.run(ctx, sessionID, input, out)
	return out
}

// run is Handle's goroutine body. Mode is resolved once here, after
// prepareRequest (the only place ApprovedProposal can flip), never inside the loop.
func (a *Agent) run(ctx context.Context, sessionID string, input TurnInput, out chan<- Event) {
	defer close(out)

	sess, err := a.store.GetSession(ctx, sessionID)
	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	implemented, err := a.store.IsSessionImplemented(ctx, sessionID)

	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	if implemented {
		a.emit(ctx, out, Event{Type: EventError, Error: "This session has finished."})
		return
	}

	lastAssistantTurn, err := a.store.LastAssistantTurn(ctx, sessionID)
	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	state := StateFromStore(sess, lastAssistantTurn)

	req, ok := a.prepareRequest(ctx, state, input, out)
	if !ok {
		return // an error event was already emitted
	}

	m := resolveMode(state.ApprovedProposal)

	vars, err := resolveModeVariables(ctx, state.ApprovedProposal, input, a.store)

	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	m.Prompt.Variables = vars

	a.runLoop(ctx, state, m, req, out)
}

// resolveMode: the flip is one-way — there is no Executing -> Planning transition.
func resolveMode(approvedProposal bool) modes.Mode {
	if approvedProposal {
		return modes.NewExecuting()
	}
	return modes.NewPlanning()
}

// resolveModeVariables builds the prompt variables a mode needs. Only Executing
// needs any (the approved plan structure); Planning gets an empty map.
func resolveModeVariables(ctx context.Context, approvedProposal bool, turn TurnInput, store store.Store) (map[string]string, error) {
	if !approvedProposal {
		return map[string]string{}, nil
	}

	p, err := store.PendingProposeStructureByCallID(ctx, turn.ToolResult.CallID)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"proposed_structure": p,
	}, nil
}
