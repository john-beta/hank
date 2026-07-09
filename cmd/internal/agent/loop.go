package agent

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/agent/phases"
	"github.com/john-beta/hank/cmd/internal/llm"
)

// maxIterations bounds the ReAct loop so a misbehaving model cannot spin forever.
const maxIterations = 10

// runLoop drives the ReAct loop: stream a turn, execute any tool calls, re-feed
// their results, and repeat until the model answers with text (no tool calls)
// or the iteration budget is exhausted. It reads and updates the caller-owned
// state (notably PrevResponseID) and exits promptly when ctx is cancelled. The
// output channel is owned and closed by run, not here.
func (a *Agent) runLoop(ctx context.Context, state *State, out chan<- Event) {
	req := llm.Request{
		Input:          state.input,
		PrevResponseID: state.PrevResponseID,
	}

	for i := 0; i < maxIterations; i++ {
		current := phases.Get(state.Phase)
		req.Instructions = current.Instructions()
		req.Tools = current.Tools()

		streamCh, err := a.llm.Stream(ctx, req)
		if err != nil {
			a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
			return
		}

		var pendingCalls []llm.FunctionCallData
		var responseID string

		if !a.consume(ctx, out, streamCh, &pendingCalls, &responseID) {
			return // ctx cancelled or a stream error was already emitted
		}

		state.recordResponse(responseID, state.Phase)

		if len(pendingCalls) == 0 {
			// Model produced a final text answer; the turn is complete.
			a.emit(ctx, out, Event{Type: EventDone, ResponseID: responseID})
			return
		}

		// Execute each tool call and prepare its result for the next turn.
		toolResults := make([]llm.ToolResult, 0, len(pendingCalls))
		for _, call := range pendingCalls {
			result, err := current.Execute(call.Name, call.Arguments)
			if err != nil {
				result = err.Error()
			}
			a.emit(ctx, out, Event{Type: EventToolResult, ToolName: call.Name, ToolResult: result})
			toolResults = append(toolResults, llm.ToolResult{CallID: call.CallID, Output: result})
		}

		state.Phase = current.Next(pendingCalls, toolResults)

		// Re-feed tool results. Instructions and Tools are re-set at the top of
		// the next iteration; the tool results are the input, so Input stays empty.
		req = llm.Request{
			PrevResponseID: responseID,
			ToolResults:    toolResults,
		}
	}

	a.emit(ctx, out, Event{Type: EventError, Error: "max iterations reached"})
}

// consume drains one stream, forwarding text/tool events and recording pending
// tool calls and the response ID. It returns false if ctx was cancelled or a
// stream error was emitted (both meaning the loop should stop), true if the
// stream ended cleanly.
func (a *Agent) consume(
	ctx context.Context,
	out chan<- Event,
	streamCh <-chan llm.StreamEvent,
	pendingCalls *[]llm.FunctionCallData,
	responseID *string,
) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case ev, ok := <-streamCh:
			if !ok {
				return true // stream closed normally
			}
			switch ev.Type {
			case "text_delta":
				a.emit(ctx, out, Event{Type: EventTextDelta, Text: ev.Text})
			case "function_call":
				if ev.FunctionCall != nil {
					a.emit(ctx, out, Event{
						Type:     EventToolCall,
						ToolName: ev.FunctionCall.Name,
						ToolArgs: ev.FunctionCall.Arguments,
					})
					*pendingCalls = append(*pendingCalls, *ev.FunctionCall)
				}
			case "done":
				*responseID = ev.ResponseID
			case "error":
				a.emit(ctx, out, Event{Type: EventError, Error: ev.Text})
				return false
			}
		}
	}
}

// emit sends an event unless ctx is cancelled first, so the loop goroutine
// never blocks on an abandoned channel.
func (a *Agent) emit(ctx context.Context, out chan<- Event, ev Event) {
	select {
	case out <- ev:
	case <-ctx.Done():
	}
}
