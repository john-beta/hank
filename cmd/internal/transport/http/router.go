package transport

import (
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

// Handler holds the dependencies shared by HTTP handlers.
type Handler struct {
	agent *agent.Agent
}

// NewRouter builds the HTTP router: the SSE messages endpoint plus the static
// placeholder at the root.
func NewRouter(ag *agent.Agent) http.Handler {
	h := &Handler{agent: ag}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/messages", h.PostMessage)
	mux.Handle("/", staticHandler())
	return mux
}
