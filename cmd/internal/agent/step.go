package agent

import (
	"context"
	"strings"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// runStep runs exactly one model response: stream it, persist the resulting
// agent turn, and then:
//   - no call        -> final text; emit done. (cont=false)
//   - one call       -> delegate to handleCall, which persists+emits the call
//     and either stops (non-auto) or executes-and-re-feeds (auto).
//
// Instructions and Tools are re-read from the mode fresh every call and never
// inherited. One call to runStep persists exactly one store.Turn; a single
// call to run (one client-visible turn) can drive several runStep calls when
// auto_re_feed tool calls chain the model into further responses without the
// client's involvement.
func (a *Agent) runStep(ctx context.Context, state *State, m mode.Interface, req llm.Request, out chan<- Event) (llm.Request, bool) {
	req.Instructions = m.Instructions()
	req.Tools = m.Tools()

	streamCh, err := a.llm.Stream(ctx, req)
	if a.fail(ctx, out, err) {
		return llm.Request{}, false
	}

	res, ok := a.consume(ctx, out, streamCh)
	if !ok {
		return llm.Request{}, false // ctx cancelled or a stream error was already emitted
	}
	state.PrevResponseID = res.responseID

	// No call: the model answered with text. Persist the agent turn and end.
	if res.call == nil {
		if _, err := a.saveAgentTurn(ctx, state, res); a.fail(ctx, out, err) {
			return llm.Request{}, false
		}
		a.emit(ctx, out, Event{Type: EventDone, ResponseID: res.responseID})
		return llm.Request{}, false
	}

	return a.handleCall(ctx, state, m, res, out)
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
// the response ID. It does not emit the tool_call event — that happens once
// auto_re_feed is known. It returns ok=false if ctx was cancelled or a stream
// error was emitted (both meaning the loop should stop).
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

// saveAgentTurn persists one agent turn produced in the loop: its streamed text
// (output_text) and response ID, read off state/res so callers never have to
// re-derive or repeat them. It returns the generated turn_id so a call emitted
// in that same response can be linked to it.
func (a *Agent) saveAgentTurn(ctx context.Context, state *State, res streamResult) (string, error) {
	respID := res.responseID
	text := res.text
	return a.store.SaveTurn(ctx, store.Turn{
		SessionID:  state.SessionID,
		Role:       "agent",
		OutputText: &text,
		ResponseID: &respID,
	})
}
