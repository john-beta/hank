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
// (intermediate tool-call turns and the final text turn), all under the
// single phase that governed the whole loop. It does not decide or persist a
// phase transition for the session — Phase is fixed for a request; advancing
// it for the next one is handled elsewhere (not yet implemented).
func (a *Agent) persistResult(ctx context.Context, state *State) {
	for _, rec := range state.turnRecords {
		a.saveAgentTurn(ctx, state.SessionID, rec.ResponseID, int(state.Phase))
	}
}

// saveAgentTurn persists a single agent turn linked to its response ID and the
// phase active when it was produced.
func (a *Agent) saveAgentTurn(ctx context.Context, sessionID, responseID string, phase int) {
	respID := responseID
	_ = a.store.SaveTurn(ctx, store.Turn{
		SessionID:  sessionID,
		Phase:      phase,
		Role:       "agent",
		ResponseID: &respID,
	})
}
