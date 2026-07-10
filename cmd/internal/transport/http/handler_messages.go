package transport

import (
	"encoding/json"
	"net/http"
)

// PostMessage handles POST /api/messages. It decodes the request, delegates to
// the agent, and streams the resulting events as SSE. Passing r.Context() to
// the agent is what wires client disconnects to loop/stream cancellation: when
// the client goes away the context cancels, the agent loop and LLM stream
// unwind, and the events channel closes, ending this handler.
func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	setSSEHeaders(w)

	events := h.agent.Handle(r.Context(), req.SessionID, req.Message)
	for event := range events {
		writeSSE(w, event)
	}
}
