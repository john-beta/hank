package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

// writeSSE flushes explicitly so the client receives each frame immediately.
func writeSSE(w http.ResponseWriter, event agent.Event) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
