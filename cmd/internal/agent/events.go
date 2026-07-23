package agent

import "context"

type EventType string

const (
	EventTextDelta  EventType = "text_delta"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventDone       EventType = "done"
	EventError      EventType = "error"
)

// Event is emitted over the streaming response. AutoReFeed is *bool so a
// meaningful false survives JSON (omitempty drops a plain false), telling the
// client it must resolve the call itself.
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

// emit sends unless ctx is cancelled first, so the goroutine never blocks on an
// abandoned channel.
func (a *Agent) emit(ctx context.Context, out chan<- Event, ev Event) {
	select {
	case out <- ev:
	case <-ctx.Done():
	}
}

// fail emits an error event and reports true if err is non-nil, so callers can
// write `if a.fail(ctx, out, err) { return ... }`.
func (a *Agent) fail(ctx context.Context, out chan<- Event, err error) bool {
	if err == nil {
		return false
	}
	a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
	return true
}
