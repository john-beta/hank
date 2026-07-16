package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// TurnInput is the discriminated input to a turn. The client sends exactly one
// of the two shapes: a plain user Message (scenario 1) or a ToolResult that
// resolves a pending non-auto call (scenario 2). Exactly one field is non-nil.
type TurnInput struct {
	Message    *string
	ToolResult *ToolResultInput
}

// ToolResultInput carries the client's resolution of a pending call. Result is
// the tool output as a JSON string.
type ToolResultInput struct {
	CallID string
	Result string
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
