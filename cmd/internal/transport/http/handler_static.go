package transport

import "net/http"

// staticHandler is a placeholder for the future web bundle (embed.FS + Vite);
// for now it reserves "/" and returns a plain-text notice.
func staticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("static files not configured"))
	})
}
