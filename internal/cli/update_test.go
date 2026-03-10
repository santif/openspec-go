package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdate_NoOpenspecDir(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "update")
	if err == nil {
		t.Fatal("expected error when no openspec directory exists")
	}
	if !strings.Contains(err.Error(), "no openspec directory") {
		t.Errorf("expected error containing 'no openspec directory', got: %v", err)
	}
}

func TestUpdate_Default(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "update")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "up to date") {
		t.Errorf("expected 'up to date' in output, got: %q", stdout)
	}
}

func TestUpdate_WithToolDir(t *testing.T) {
	root := setupProject(t)

	// Create a .claude directory to simulate tool presence
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	stdout, _, err := executeCommand(t, root, "update")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With a tool dir present, should report updating or up to date
	if !strings.Contains(stdout, "Updated") && !strings.Contains(stdout, "up to date") {
		t.Errorf("expected 'Updated' or 'up to date' in output, got: %q", stdout)
	}
}
