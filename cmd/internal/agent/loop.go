package agent

import (
	"context"
	"strings"

	"github.com/john-beta/hank/cmd/internal/agent/modes"
	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// maxIterations bounds the ReAct loop so a misbehaving model cannot spin forever.
const maxIterations = 10

// runLoop drives the ReAct loop for a fixed mode. ParallelToolCalls is off, so
// each model response yields at most one function call. Per iteration it streams
// a turn, persists the resulting agent turn, and then:
//   - no call        -> the model produced final text; emit done and return.
//   - one auto call  -> execute the stub, record its result, re-feed, continue.
//   - one non-auto   -> emit the enriched tool_call event and stop the turn; the
//     client resolves it in the next request.
//
// The initial req is supplied by run (either fresh input or a client tool
// result). Instructions and Tools are re-read from the mode fresh every
// iteration and never inherited. The output channel is owned and closed by run.
func (a *Agent) runLoop(ctx context.Context, state *State, modeID mode.ID, req llm.Request, out chan<- Event) {
	m := modes.Get(modeID)

	for i := 0; i < maxIterations; i++ {
		req.Instructions = m.Instructions()
		req.Tools = m.Tools()

		streamCh, err := a.llm.Stream(ctx, req)
		if err != nil {
			a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
			return
		}

		res, ok := a.consume(ctx, out, streamCh)
		if !ok {
			return // ctx cancelled or a stream error was already emitted
		}
		state.PrevResponseID = res.responseID

		// No call: the model answered with text. Persist the agent turn and end.
		if res.call == nil {
			if _, err := a.saveAgentTurn(ctx, state.SessionID, res.responseID, res.text); err != nil {
				a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
				return
			}
			a.emit(ctx, out, Event{Type: EventDone, ResponseID: res.responseID})
			return
		}

		// One call. Persist the agent turn that produced it, then INSERT the call
		// row (result NULL) linked to that turn.
		call := res.call
		auto := mode.AutoReFeed(m.Tools(), call.Name)

		turnID, err := a.saveAgentTurn(ctx, state.SessionID, res.responseID, res.text)
		if err != nil {
			a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
			return
		}
		args := call.Arguments
		if err := a.store.SaveCall(ctx, store.Call{
			CallID:     call.CallID,
			TurnID:     turnID,
			Name:       call.Name,
			Args:       &args,
			AutoReFeed: auto,
		}); err != nil {
			a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
			return
		}

		// The tool_call event carries auto_re_feed now that it is known.
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
			return
		}

		// Auto: execute the (stub) tool, record its result, and re-feed. The
		// client's tool result travels this same Request shape — only the origin
		// of the output differs.
		output, execErr := m.Execute(call.Name, call.Arguments)
		if execErr != nil {
			output = execErr.Error()
		}
		if err := a.store.UpdateCallResult(ctx, call.CallID, output); err != nil {
			a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
			return
		}
		a.emit(ctx, out, Event{Type: EventToolResult, ToolName: call.Name, ToolResult: output})

		req = llm.Request{
			ToolResults:    []llm.ToolResult{{CallID: call.CallID, Output: output}},
			PrevResponseID: res.responseID,
		}
	}

	a.emit(ctx, out, Event{Type: EventError, Error: "max iterations reached"})
}

// streamResult is the outcome of consuming one model response: the single
// assembled function call (nil if the model produced only text), the response
// ID, and the concatenated streamed text.
type streamResult struct {
	call       *llm.FunctionCallData
	responseID string
	text       string
}

// consume drains one stream, forwarding cosmetic text deltas and accumulating
// the streamed text, the single function call (ParallelToolCalls is off), and
// the response ID. It does not emit the tool_call event — that happens in the
// loop once auto_re_feed is known. It returns ok=false if ctx was cancelled or a
// stream error was emitted (both meaning the loop should stop).
func (a *Agent) consume(ctx context.Context, out chan<- Event, streamCh <-chan llm.StreamEvent) (streamResult, bool) {
	var res streamResult
	var text strings.Builder
	for {
		select {
		case <-ctx.Done():
			return res, false
		case ev, okCh := <-streamCh:
			if !okCh {
				res.text = text.String()
				return res, true // stream closed normally
			}
			switch ev.Type {
			case "text_delta":
				text.WriteString(ev.Text)
				a.emit(ctx, out, Event{Type: EventTextDelta, Text: ev.Text})
			case "function_call":
				if ev.FunctionCall != nil {
					res.call = ev.FunctionCall
				}
			case "done":
				res.responseID = ev.ResponseID
			case "error":
				a.emit(ctx, out, Event{Type: EventError, Error: ev.Text})
				return res, false
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
