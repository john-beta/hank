package agent

import (
	"context"
	"strings"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// runStep runs exactly one model response and either ends the turn (no call,
// or a non-auto call) or hands off to handleCall for an auto one. Instructions
// and Tools are re-read from the mode on every call, never cached across
// iterations — OpenAI's PreviousResponseID does not carry them forward, so a
// cached copy would silently drop instructions/tools after the first Step.
// One runStep call persists exactly one store.Turn; a single client-visible
// turn can drive several Steps when auto_re_feed chains the model further
// without the client's involvement (see CLAUDE.md, "Turn vs. Step").
func (a *Agent) runStep(ctx context.Context, state *State, m mode.Mode, req llm.Request, out chan<- Event) (llm.Request, bool) {
	req.Instructions = m.Instructions
	req.Tools = m.Tools

	streamCh, err := a.llm.Stream(ctx, req)
	if a.fail(ctx, out, err) {
		return llm.Request{}, false
	}

	res, ok := a.consume(ctx, out, streamCh)
	if !ok {
		return llm.Request{}, false // ctx cancelled or a stream error was already emitted
	}
	state.PrevResponseID = res.responseID

	if res.call == nil {
		if _, err := a.saveAgentTurn(ctx, state, res); a.fail(ctx, out, err) {
			return llm.Request{}, false
		}
		a.emit(ctx, out, Event{Type: EventDone, ResponseID: res.responseID})
		return llm.Request{}, false
	}

	return a.handleCall(ctx, state, m, res, out)
}

type streamResult struct {
	call       *llm.FunctionCallData
	responseID string
	text       string
}

// consume drains one stream: forwards text deltas as they arrive, accumulates
// the single function call (ParallelToolCalls is off, so there's never more
// than one), and does not emit tool_call itself — handleCall does, once
// auto_re_feed is known. ok=false means ctx was cancelled or a stream error
// was already emitted; either way the loop should stop.
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

// saveAgentTurn reads sessionID/responseID/text off state/res so callers
// never re-derive or repeat them.
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
