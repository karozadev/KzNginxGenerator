package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetGenerateFlags() {
	genDomain = ""
	genProxy = ""
	genOut = ""
	genSSLCert = ""
	genSSLKey = ""
	genWebSocket = false
}

func runRoot(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	resetGenerateFlags()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return buf.String(), err
}

func TestGenerateCmd_StdoutOutput(t *testing.T) {
	out, err := runRoot(t, "generate", "--domain=example.com", "--proxy=http://localhost:8000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "server_name example.com;") {
		t.Errorf("expected output to contain server_name directive, got:\n%s", out)
	}
	if !strings.Contains(out, "proxy_pass http://localhost:8000;") {
		t.Errorf("expected output to contain proxy_pass directive, got:\n%s", out)
	}
}

func TestGenerateCmd_WebSocketAndSSL(t *testing.T) {
	out, err := runRoot(t, "generate",
		"--domain=example.com",
		"--proxy=http://localhost:8000",
		"--websocket",
		"--ssl-cert=/etc/ssl/example.com.crt",
		"--ssl-key=/etc/ssl/example.com.key",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ssl_certificate /etc/ssl/example.com.crt;") {
		t.Errorf("expected ssl_certificate directive, got:\n%s", out)
	}
	if !strings.Contains(out, "proxy_set_header Upgrade $http_upgrade;") {
		t.Errorf("expected websocket upgrade header, got:\n%s", out)
	}
}

func TestGenerateCmd_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "karoza-app.conf")

	_, err := runRoot(t, "generate", "--domain=example.com", "--proxy=http://localhost:8000", "--out="+outPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if !strings.Contains(string(data), "server_name example.com;") {
		t.Errorf("expected file content to contain server_name directive, got:\n%s", data)
	}
}

func TestGenerateCmd_MissingDomain(t *testing.T) {
	_, err := runRoot(t, "generate", "--proxy=http://localhost:8000")
	if err == nil {
		t.Fatal("expected error when --domain is missing")
	}
}

func TestGenerateCmd_MissingProxy(t *testing.T) {
	_, err := runRoot(t, "generate", "--domain=example.com")
	if err == nil {
		t.Fatal("expected error when --proxy is missing")
	}
}
