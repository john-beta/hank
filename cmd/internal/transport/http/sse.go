package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

// setSSEHeaders configures the response for Server-Sent Events.
func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

// writeSSE serialises an agent event as one SSE "data:" frame and flushes it so
// the client receives it immediately.
func writeSSE(w http.ResponseWriter, event agent.Event) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
