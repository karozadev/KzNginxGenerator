// Package server implements the lightweight local HTTP server that
// backs the KzNginxGenerator Web UI: it serves the embedded static
// assets and exposes a small JSON API used to render Nginx
// configurations in real time.
package server

import (
	"net/http"

	"github.com/karoza/kz-nginx-generator/web"
)

// Options configures the HTTP server.
type Options struct {
	// Addr is the address to listen on, e.g. ":8080".
	Addr string
	// Version is the kznginx build version, surfaced via /api/version.
	Version string
	// Revision is the git commit the binary was built from, surfaced
	// via /api/version. May be empty.
	Revision string
}

// New builds an *http.Server ready to ListenAndServe, wired with the
// KzNginxGenerator Web UI and API routes.
func New(opts Options) *http.Server {
	return &http.Server{
		Addr:    opts.Addr,
		Handler: NewMux(opts),
	}
}

// NewMux builds the http.Handler serving the Web UI and API routes,
// without binding it to a listen address. Exposed separately from New
// so tests can exercise routes with httptest without starting a real
// network listener.
func NewMux(opts Options) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(web.FS)))
	mux.HandleFunc("/api/generate", handleGenerate)
	mux.HandleFunc("/api/version", handleVersion(opts))
	mux.HandleFunc("/healthz", handleHealth)

	return mux
}
