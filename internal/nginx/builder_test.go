package nginx

import (
	"strings"
	"testing"
)

func TestNewSimpleReverseProxyConfig_PlainHTTP(t *testing.T) {
	cfg := NewSimpleReverseProxyConfig(SimpleReverseProxyOptions{
		Domain:    "example.com",
		ProxyPass: "http://localhost:8000",
	})
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
	out, err := Render(cfg)
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	if !strings.Contains(out, "server_name example.com;") {
		t.Errorf("expected server_name directive, got:\n%s", out)
	}
	if !strings.Contains(out, "proxy_pass http://localhost:8000;") {
		t.Errorf("expected proxy_pass directive, got:\n%s", out)
	}
	if strings.Contains(out, "ssl_certificate") {
		t.Errorf("expected no SSL directives for plain HTTP config, got:\n%s", out)
	}
}

func TestNewSimpleReverseProxyConfig_WithSSL(t *testing.T) {
	cfg := NewSimpleReverseProxyConfig(SimpleReverseProxyOptions{
		Domain:                "example.com",
		ProxyPass:             "http://localhost:8000",
		WebSocket:             true,
		SSLCertificatePath:    "/etc/ssl/example.com.crt",
		SSLCertificateKeyPath: "/etc/ssl/example.com.key",
	})
	out, err := Render(cfg)
	if err != nil {
		t.Fatalf("Render() unexpected error: %v", err)
	}
	if !strings.Contains(out, "ssl_certificate /etc/ssl/example.com.crt;") {
		t.Errorf("expected ssl_certificate directive, got:\n%s", out)
	}
	if !strings.Contains(out, "return 301 https://$host$request_uri;") {
		t.Errorf("expected HTTP->HTTPS redirect, got:\n%s", out)
	}
	if !strings.Contains(out, "proxy_set_header Upgrade $http_upgrade;") {
		t.Errorf("expected websocket upgrade header, got:\n%s", out)
	}
}
