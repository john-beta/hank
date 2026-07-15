package transport

import (
	"encoding/json"
	"net/http"
)

// CreateSession handles POST /api/sessions. It reads the workspace root_dir,
// creates a new session directly in the store (no LLM involvement), and returns
// its ID as JSON.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.RootDir == "" {
		http.Error(w, "root_dir is required", http.StatusBadRequest)
		return
	}

	sess, err := h.store.CreateSession(r.Context(), req.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateSessionResponse{SessionID: sess.SessionID})
}
