package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/karoza/kz-nginx-generator/internal/nginx"
)

func testMux() http.Handler {
	return NewMux(Options{Version: "v1.2.3-beta.1", Revision: "abc1234"})
}

func TestHandleGenerate_Valid(t *testing.T) {
	cfg := nginx.Config{
		Servers: []nginx.Server{
			{
				ServerNames: []string{"example.com"},
				Locations: []nginx.Location{
					{Path: "/", ProxyPass: "http://127.0.0.1:8000"},
				},
			},
		},
	}
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp generateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !strings.Contains(resp.Output, "server_name example.com;") {
		t.Errorf("expected output to contain server_name directive, got: %s", resp.Output)
	}
	if !strings.Contains(resp.Output, "proxy_pass http://127.0.0.1:8000;") {
		t.Errorf("expected output to contain proxy_pass directive, got: %s", resp.Output)
	}
}

func TestHandleGenerate_InvalidConfig(t *testing.T) {
	body := []byte(`{"servers": []}`)

	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp generateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestHandleGenerate_MalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleGenerate_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/generate", nil)
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp versionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Version != "v1.2.3-beta.1" {
		t.Errorf("expected version v1.2.3-beta.1, got %q", resp.Version)
	}
	if resp.Revision != "abc1234" {
		t.Errorf("expected revision abc1234, got %q", resp.Revision)
	}
}

func TestHandleVersion_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/version", nil)
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	testMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body \"ok\", got %q", rec.Body.String())
	}
}

func TestStaticAssets(t *testing.T) {
	tests := []struct {
		path     string
		contains string
	}{
		{"/", "KzNginxGenerator"},
		{"/static/app.js", "scheduleGenerate"},
		{"/static/style.css", "code-block"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			testMux().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tt.path, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tt.contains) {
				t.Errorf("expected %s body to contain %q", tt.path, tt.contains)
			}
		})
	}
}
