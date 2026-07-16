package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// run is the persistence wrapper around the pure ReAct loop: load session state,
// perform any pre-loop mode transition, record the user turn, build the initial
// LLM request, resolve the mode once, then run the loop. It owns the output
// channel and always closes it.
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

	// Build the initial request. This is where the (possible) Planning ->
	// Executing transition happens — before the loop, never inside it.
	req, ok := a.prepareRequest(ctx, state, input, out)
	if !ok {
		return // an error event was already emitted
	}

	// Mode is derived from the (possibly just-flipped) approval boolean, resolved
	// once and held fixed for the whole loop.
	modeID := mode.Resolve(state.ApprovedProposal)

	a.runLoop(ctx, state, modeID, req, out)
}

// prepareRequest records the user turn and constructs the initial llm.Request
// for this turn. For a plain Message it is a fresh-input request; for a
// ToolResult it resolves the pending call, applies the approval flip if the
// result approves, and builds a tool-result request — the exact same Request
// shape the loop uses to re-feed auto tools (one Request type, two sources). It
// returns false (after emitting an error) if the turn cannot proceed.
func (a *Agent) prepareRequest(ctx context.Context, state *State, input TurnInput, out chan<- Event) (llm.Request, bool) {
	switch {
	case input.ToolResult != nil:
		return a.prepareToolResultRequest(ctx, state, input.ToolResult, out)
	case input.Message != nil:
		return a.prepareMessageRequest(ctx, state, *input.Message, out)
	default:
		a.emit(ctx, out, Event{Type: EventError, Error: "agent: turn input has neither message nor tool result"})
		return llm.Request{}, false
	}
}

// prepareMessageRequest handles scenario 1: a plain user message. The message is
// persisted as the user turn's output_text and seeds the request input.
func (a *Agent) prepareMessageRequest(ctx context.Context, state *State, message string, out chan<- Event) (llm.Request, bool) {
	msg := message
	if _, err := a.store.SaveTurn(ctx, store.Turn{
		SessionID:  state.SessionID,
		Role:       "user",
		OutputText: &msg,
	}); err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return llm.Request{}, false
	}
	return llm.Request{Input: message, PrevResponseID: state.PrevResponseID}, true
}

// prepareToolResultRequest handles scenario 2: the client resolving a pending
// non-auto call. It validates the incoming call against the store's pending
// call, records the result, applies the approval flip, persists the user turn
// (output_text NULL — the payload lives in the call.result UPDATE), and builds
// the tool-result request.
func (a *Agent) prepareToolResultRequest(ctx context.Context, state *State, tr *ToolResultInput, out chan<- Event) (llm.Request, bool) {
	pending, err := a.store.PendingCall(ctx, state.SessionID)
	if err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return llm.Request{}, false
	}
	if pending == nil || pending.CallID != tr.CallID {
		// Do not forward a desynced call to OpenAI.
		a.emit(ctx, out, Event{Type: EventError, Error: fmt.Sprintf("agent: tool result call_id %q does not match the pending call", tr.CallID)})
		return llm.Request{}, false
	}

	if err := a.store.UpdateCallResult(ctx, tr.CallID, tr.Result); err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return llm.Request{}, false
	}

	// Read a single `approved` boolean from the result JSON — no other
	// interpretation of the payload. An approving result flips this in-memory
	// state to Executing for the rest of this turn; nothing is persisted, so
	// the next request starts back at Planning unless it too carries an
	// approving result.
	if approvedFromResult(tr.Result) {
		state.ApprovedProposal = true
	}

	if _, err := a.store.SaveTurn(ctx, store.Turn{
		SessionID: state.SessionID,
		Role:      "user",
		// output_text is NULL: the tool-result payload lives in call.result.
	}); err != nil {
		a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
		return llm.Request{}, false
	}

	return llm.Request{
		ToolResults:    []llm.ToolResult{{CallID: tr.CallID, Output: tr.Result}},
		PrevResponseID: state.PrevResponseID,
	}, true
}

// saveAgentTurn persists one agent turn produced in the loop: its streamed text
// (output_text) and response ID. It returns the generated turn_id so a call
// emitted in that same response can be linked to it.
func (a *Agent) saveAgentTurn(ctx context.Context, sessionID, responseID, outputText string) (string, error) {
	respID := responseID
	text := outputText
	return a.store.SaveTurn(ctx, store.Turn{
		SessionID:  sessionID,
		Role:       "agent",
		OutputText: &text,
		ResponseID: &respID,
	})
}

// approvedFromResult extracts the single `approved` boolean from a tool result's
// JSON. A malformed or approval-less payload reports false.
func approvedFromResult(result string) bool {
	var payload struct {
		Approved bool `json:"approved"`
	}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return false
	}
	return payload.Approved
}
