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

// Event is JSON-serialisable for direct emission over SSE. tool_call is the
// only non-cosmetic type: an AutoReFeed of false signals the client that it
// must resolve the call itself. AutoReFeed is a *bool, not bool, so the
// meaningful "false" survives JSON encoding — a plain bool would be omitted.
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

// emit sends unless ctx is cancelled first, so the loop goroutine never
// blocks on an abandoned channel.
func (a *Agent) emit(ctx context.Context, out chan<- Event, ev Event) {
	select {
	case out <- ev:
	case <-ctx.Done():
	}
}

// fail emits an error event and reports true if err is non-nil, so callers
// can write `if a.fail(ctx, out, err) { return ... }`. Used by loop.go,
// step.go and call.go; input.go keeps its own explicit emit calls instead.
func (a *Agent) fail(ctx context.Context, out chan<- Event, err error) bool {
	if err == nil {
		return false
	}
	a.emit(ctx, out, Event{Type: EventError, Error: err.Error()})
	return true
}
