package transport

import (
	"encoding/json"
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

type StreamingJSONError struct {
	Type  string
	Error string
}

func setStreamingHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
}

// writeStreaming flushes explicitly so the client receives each frame immediately.
// Each event is written as its own JSON line, per the NDJSON format.
func writeStreaming(w http.ResponseWriter, event agent.Event) {
	data, err := json.Marshal(event)

	if err != nil {
		data, _ = json.Marshal(StreamingJSONError{Type: "error", Error: err.Error()})
	}

	w.Write(data)
	w.Write([]byte("\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
