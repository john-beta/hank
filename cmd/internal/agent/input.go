package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// TurnInput is a discriminated union: exactly one of Message or ToolResult is
// non-nil.
type TurnInput struct {
	Message    *string
	ToolResult *ToolResultInput
}

// ToolResultInput carries the client's resolution of a pending call; Result is
// the tool output as a raw JSON string.
type ToolResultInput struct {
	CallID string
	Result string
}

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

	// An approving result flips ApprovedProposal to true in memory, for the
	// rest of this request only. Nothing is persisted, so the next request
	// starts back at Planning unless it too carries an approving result.
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

// approvedFromResult reports the result JSON's `approved` field. A malformed
// or approval-less payload reports false, not an error — approval is opt-in.
func approvedFromResult(result string) bool {
	var payload struct {
		Approved bool `json:"approved"`
	}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return false
	}
	return payload.Approved
}
