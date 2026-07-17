package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// handleCall persists the turn and call row (result NULL), then either stops
// for the client to resolve a non-auto call or executes and re-feeds an auto
// one. tool_call is emitted here, not in consume, because auto_re_feed isn't
// known until the call is matched against the mode's tools.
func (a *Agent) handleCall(ctx context.Context, state *State, m mode.Mode, res streamResult, out chan<- Event) (llm.Request, bool) {
	call := res.call
	auto := m.AutoReFeed(call.Name)

	turnID, err := a.saveAgentTurn(ctx, state, res)
	if a.fail(ctx, out, err) {
		return llm.Request{}, false
	}

	args := call.Arguments
	if err := a.store.SaveCall(ctx, store.Call{
		CallID:     call.CallID,
		TurnID:     turnID,
		Name:       call.Name,
		Args:       &args,
		AutoReFeed: auto,
	}); a.fail(ctx, out, err) {
		return llm.Request{}, false
	}

	autoVal := auto
	a.emit(ctx, out, Event{
		Type:       EventToolCall,
		ToolName:   call.Name,
		ToolArgs:   call.Arguments,
		CallID:     call.CallID,
		AutoReFeed: &autoVal,
	})

	if !auto {
		return llm.Request{}, false // client resolves it; the loop must not spin
	}

	output, execErr := m.Execute(call.Name, call.Arguments)
	if execErr != nil {
		output = execErr.Error()
	}
	if err := a.store.UpdateCallResult(ctx, call.CallID, output); a.fail(ctx, out, err) {
		return llm.Request{}, false
	}
	a.emit(ctx, out, Event{Type: EventToolResult, ToolName: call.Name, ToolResult: output})

	return llm.Request{
		ToolResults:    []llm.ToolResult{{CallID: call.CallID, Output: output}},
		PrevResponseID: res.responseID,
	}, true
}
