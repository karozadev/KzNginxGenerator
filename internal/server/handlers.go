package server

import (
	"encoding/json"
	"net/http"

	"github.com/karoza/kz-nginx-generator/internal/nginx"
)

type generateResponse struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// handleGenerate accepts a JSON-encoded nginx.Config in the request
// body and responds with the rendered Nginx configuration text.
func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, generateResponse{Error: "method not allowed, use POST"})
		return
	}

	var cfg nginx.Config
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, generateResponse{Error: "invalid JSON body: " + err.Error()})
		return
	}

	output, err := nginx.Render(cfg)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, generateResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, generateResponse{Output: output})
}

type versionResponse struct {
	Version  string `json:"version"`
	Revision string `json:"revision,omitempty"`
}

// handleVersion responds with the running kznginx build's version and
// git revision.
func handleVersion(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, versionResponse{Version: opts.Version, Revision: opts.Revision})
	}
}

// handleHealth is a trivial liveness probe endpoint.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
