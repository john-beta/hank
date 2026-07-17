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

func NewRouter(ag *agent.Agent, st store.Store) http.Handler {
	h := &Handler{agent: ag, store: st}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/sessions", h.CreateSession)
	mux.HandleFunc("POST /api/messages", h.PostMessage)
	mux.Handle("/", staticHandler())
	return mux
}
