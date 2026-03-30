package web

import (
	"net/http"

	"github.com/miqdadyyy/flaggo"
)

// Handler returns an http.Handler that serves the dashboard UI and proxies
// JSON API requests to flaggo's built-in handler.
//
// - Non-JSON GET requests serve the embedded dashboard HTML.
// - All JSON requests (Accept: application/json) are handled by ff.GetHandler().
//
// If the Flaggo instance was created with Config.Username/Password,
// Basic Authentication is applied to both the dashboard and API.
func Handler(ff *flaggo.Flaggo) http.Handler {
	return ff.WrapAuth(&uiHandler{api: ff.GetHandler()})
}

type uiHandler struct {
	api http.Handler
}

func (h *uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == "application/json" {
		h.api.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(indexView))
}
