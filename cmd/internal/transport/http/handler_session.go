package transport

import (
	"encoding/json"
	"net/http"
)

// CreateSession handles POST /api/sessions. It creates a new session directly in
// the store (no LLM involvement) and returns its ID as JSON.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.CreateSession(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateSessionResponse{SessionID: sess.SessionID})
}
