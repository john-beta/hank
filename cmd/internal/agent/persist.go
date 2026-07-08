package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/store"
)

// run is the persistence wrapper around the pure ReAct loop: load session state,
// record the user turn, run the loop, then persist the resulting agent turn and
// phase. It owns the output channel and always closes it.
func (a *Agent) run(ctx context.Context, sessionID, userText string, out chan<- Event) {
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
	state.input = userText

	if err := a.store.SaveTurn(ctx, store.Turn{
		SessionID: sessionID,
		Phase:     int(state.Phase),
		Role:      "user",
		Content:   &userText,
	}); err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return
	}

	a.runLoop(ctx, state, out)

	a.persistResult(ctx, state)
}

// persistResult records one agent turn per model response produced in the loop
// (intermediate tool-call turns and the final text turn) and the session's
// current phase after the loop completes.
func (a *Agent) persistResult(ctx context.Context, state *State) {
	for _, respID := range state.responseIDs {
		a.saveAgentTurn(ctx, state, respID)
	}
	_ = a.store.UpdateSessionPhase(ctx, state.SessionID, int(state.Phase))
}

// saveAgentTurn persists a single agent turn linked to its response ID.
func (a *Agent) saveAgentTurn(ctx context.Context, state *State, responseID string) {
	respID := responseID
	_ = a.store.SaveTurn(ctx, store.Turn{
		SessionID:  state.SessionID,
		Phase:      int(state.Phase),
		Role:       "agent",
		ResponseID: &respID,
	})
}
