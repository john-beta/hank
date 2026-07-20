package transport

import (
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
	"github.com/john-beta/hank/cmd/internal/store"
)

// Handler holds the dependencies shared by HTTP handlers.
type Handler struct {
	agent *agent.Agent
	store store.Store
}

const allowedOrigin = "http://localhost:5173"

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(ag *agent.Agent, st store.Store) http.Handler {
	h := &Handler{agent: ag, store: st}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions", h.CreateSession)
	mux.HandleFunc("POST /api/messages", h.PostMessage)
	mux.Handle("/", staticHandler())
	return corsMiddleware(mux)
}
