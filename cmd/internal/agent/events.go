package agent

import "context"

// EventType enumerates the agent-level events streamed to transports.
type EventType string

const (
	EventTextDelta  EventType = "text_delta"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventDone       EventType = "done"
	EventError      EventType = "error"
)

// Event is a single item in the agent's output stream. It is JSON-serialisable
// for direct emission over SSE.
//
// The tool_call event is the only non-cosmetic one: it carries CallID and
// AutoReFeed, and an AutoReFeed of false signals the client that it must
// construct and send the next request as a tool result. text_delta and
// tool_result are cosmetic. AutoReFeed is a *bool so that the meaningful
// "false" value survives JSON encoding (a plain bool would be omitted).
type Event struct {
	Type       EventType `json:"type"`
	Text       string    `json:"text,omitempty"`
	ToolName   string    `json:"tool_name,omitempty"`
	ToolArgs   string    `json:"tool_args,omitempty"`
	ToolResult string    `json:"tool_result,omitempty"`
	CallID     string    `json:"call_id,omitempty"`
	AutoReFeed *bool     `json:"auto_re_feed,omitempty"`
	Error      string    `json:"error,omitempty"`
	ResponseID string    `json:"response_id,omitempty"`
}

// emit sends an event unless ctx is cancelled first, so the loop goroutine
// never blocks on an abandoned channel.
func (a *Agent) emit(ctx context.Context, out chan<- Event, ev Event) {
	select {
	case out <- ev:
	case <-ctx.Done():
	}
}

// fail emits an error event and reports true if err is non-nil, so callers in
// the ReAct loop can write `if a.fail(ctx, out, err) { return ... }` instead of
// repeating the emit. Scoped to loop.go/step.go/call.go for now — persist.go
// keeps its own explicit emit calls.
func (a *Agent) fail(ctx context.Context, out chan<- Event, err error) bool {
	if err == nil {
		return false
	}
	a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
	return true
}
