package transport

import (
	"encoding/json"
	"net/http"
	"os"
)

// CreateSession creates the session directly in the store — no LLM involvement.
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

	_, err := os.Stat(req.RootDir)
	if os.IsNotExist(err) {
		http.Error(w, "root_dir does not exist", http.StatusNotFound)
		return
	}

	sess, err := h.store.CreateSession(r.Context(), req.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateSessionResponse{SessionID: sess.SessionID, RootDir: sess.RootDir, CreatedAt: sess.CreatedAt})
}
