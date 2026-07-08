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
	tools *ToolRegistry
}

// New wires an agent with its LLM client and store.
func New(llmClient llm.Client, st store.Store) *Agent {
	return &Agent{
		llm:   llmClient,
		store: st,
		tools: NewToolRegistry(),
	}
}

// Handle starts a turn for userText within sessionID and returns a channel of
// events. The channel is closed when the turn completes or ctx is cancelled.
func (a *Agent) Handle(ctx context.Context, sessionID, userText string) <-chan Event {
	out := make(chan Event)
	go a.run(ctx, sessionID, userText, out)
	return out
}

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
