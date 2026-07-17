package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// handleCall runs once a step's response includes a call: it persists the
// turn and the call row (result NULL), then either stops for the client to
// resolve a non-auto call, or executes and re-feeds an auto one. The
// tool_call event is emitted here rather than in consume/step because
// auto_re_feed isn't known until the call is matched against the mode's tools.
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
		// Hand the call back to the client; the loop must not spin waiting.
		return llm.Request{}, false
	}

	// Client-sourced and loop-sourced tool results travel this same Request
	// shape — only the origin of the output differs.
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
