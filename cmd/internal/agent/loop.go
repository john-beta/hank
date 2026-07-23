package agent

import (
	"context"
	"strings"

	"github.com/john-beta/hank/cmd/internal/agent/modes"
	"github.com/john-beta/hank/cmd/internal/llm"
	"github.com/john-beta/hank/cmd/internal/store"
)

// maxIterations bounds the loop so a misbehaving model cannot spin forever.
const maxIterations = 15

// runLoop drives the bounded ReAct loop. runStep emits its own terminal events
// before reporting cont=false, so runLoop's only job then is to stop.
func (a *Agent) runLoop(ctx context.Context, state *State, m modes.Mode, req llm.Request, out chan<- Event) {
	for i := 0; i < maxIterations; i++ {
		next, cont := a.runStep(ctx, state, m, req, out)
		if !cont {
			return
		}
		req = next
	}

	a.emit(ctx, out, Event{Type: EventError, Error: "max iterations reached"})
}

// runStep runs one model response and either ends the turn or hands an auto
// call off to handleCall. The Prompt is re-set every Step because OpenAI's
// PreviousResponseID does not carry it forward.
func (a *Agent) runStep(ctx context.Context, state *State, m modes.Mode, req llm.Request, out chan<- Event) (llm.Request, bool) {
	req.Prompt = m.Prompt

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
		if _, err := a.saveAssistantTurn(ctx, state, res); a.fail(ctx, out, err) {
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

// consume drains one stream: forwards text deltas, accumulates the single
// function call (ParallelToolCalls is off, so never more than one), and does
// not emit tool_call itself — handleCall does, once auto_re_feed is known.
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

// handleCall persists the turn and call row (result NULL), then either stops
// for the client to resolve a non-auto call or executes and re-feeds an auto
// one. tool_call is emitted here, not in consume, because auto_re_feed isn't
// known until the call is matched against the mode's tools.
func (a *Agent) handleCall(ctx context.Context, state *State, m modes.Mode, res streamResult, out chan<- Event) (llm.Request, bool) {
	call := res.call
	auto := m.AutoReFeed(call.Name)

	turnID, err := a.saveAssistantTurn(ctx, state, res)
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

	output, ctxErr := m.Execute(ctx, call.Name, call.Arguments, modes.ExecState{
		SessionID: state.SessionID,
		RootDir:   state.RootDir,
	})
	if ctxErr != nil {
		return llm.Request{}, false
	}
	if err := a.store.UpdateCallResult(ctx, call.CallID, output); a.fail(ctx, out, err) {
		return llm.Request{}, false
	}
	a.emit(ctx, out, Event{Type: EventToolResult, CallID: call.CallID, ToolName: call.Name, ToolResult: output})

	return llm.Request{
		ToolResults:    []llm.ToolResult{{CallID: call.CallID, Output: output}},
		PrevResponseID: res.responseID,
	}, true
}

func (a *Agent) saveAssistantTurn(ctx context.Context, state *State, res streamResult) (string, error) {
	respID := res.responseID
	text := res.text
	return a.store.SaveTurn(ctx, store.Turn{
		SessionID:  state.SessionID,
		Role:       "assistant",
		OutputText: &text,
		ResponseID: &respID,
	})
}
