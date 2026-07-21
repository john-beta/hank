package transport

import (
	"encoding/json"
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

type SSEJsonError struct {
	Type  string
	Error string
}

func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

// writeSSE flushes explicitly so the client receives each frame immediately.
func writeSSE(w http.ResponseWriter, event agent.Event) {
	data, err := json.Marshal(event)

	if err != nil {
		data, _ = json.Marshal(SSEJsonError{Type: "error", Error: err.Error()})
	}

	w.Write(data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
