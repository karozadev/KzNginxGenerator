package cmd

import (
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	prevVersion, prevCommit := Version, GitCommit
	Version = "v1.2.3-beta.1"
	GitCommit = "abc1234"
	defer func() { Version, GitCommit = prevVersion, prevCommit }()

	out, err := runRoot(t, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "v1.2.3-beta.1") {
		t.Errorf("expected output to contain version, got: %q", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("expected output to contain git revision, got: %q", out)
	}
}

func TestVersionCmd_NoRevision(t *testing.T) {
	prevVersion, prevCommit := Version, GitCommit
	Version = "v1.0.0"
	GitCommit = "none"
	defer func() { Version, GitCommit = prevVersion, prevCommit }()

	out, err := runRoot(t, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "v1.0.0") {
		t.Errorf("expected output to contain version, got: %q", out)
	}
	if strings.Contains(out, "(none)") {
		t.Errorf("expected no revision suffix when GitCommit is \"none\", got: %q", out)
	}
}
