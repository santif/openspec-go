package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- instructions command ---

func TestInstructions_NoArgs_ListsArtifacts(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Available artifacts:") {
		t.Errorf("expected 'Available artifacts:' in output, got: %q", stdout)
	}
	// Should list artifact names from the spec-driven schema
	for _, name := range []string{"proposal", "spec", "design", "tasks"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("expected artifact %q in output, got: %q", name, stdout)
		}
	}
}

func TestInstructions_Proposal(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions", "proposal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty instruction text for proposal")
	}
}

func TestInstructions_ProposalJSON(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions", "proposal", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %q", jsonErr, stdout)
	}
	if _, ok := result["artifactId"]; !ok {
		t.Error("expected 'artifactId' field in JSON output")
	}
	if _, ok := result["instruction"]; !ok {
		t.Error("expected 'instruction' field in JSON output")
	}
}

func TestInstructions_Apply(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions", "apply")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty instruction text for apply")
	}
}

func TestInstructions_ApplyJSON(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions", "apply", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %q", jsonErr, stdout)
	}
	if result["artifactId"] != "apply" {
		t.Errorf("expected artifactId 'apply', got %v", result["artifactId"])
	}
}

func TestInstructions_InvalidArtifact(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "instructions", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent artifact")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error containing 'not found', got: %v", err)
	}
}

func TestInstructions_SchemaOverride(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "instructions", "--schema", "spec-driven", "proposal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected non-empty instruction text with schema override")
	}
}

// --- view command ---

func TestView_NoOpenspecDir(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "view")
	if err == nil {
		t.Error("expected error when no openspec dir exists")
	}
	if err != nil && !strings.Contains(err.Error(), "openspec") {
		t.Errorf("expected error mentioning 'openspec', got: %v", err)
	}
}

func TestView_Empty(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "view")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Dashboard") {
		t.Errorf("expected 'Dashboard' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "(none)") {
		t.Errorf("expected '(none)' markers in output, got: %q", stdout)
	}
}

func TestView_WithSpecs(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "view")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "user-auth") {
		t.Errorf("expected 'user-auth' in output, got: %q", stdout)
	}
}

func TestView_WithChanges(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "view")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "add-auth") {
		t.Errorf("expected 'add-auth' in output, got: %q", stdout)
	}
	// Should contain progress indicators (e.g., counts like "1/4" or similar)
	if !strings.Contains(stdout, "/") {
		t.Errorf("expected progress indicators in output, got: %q", stdout)
	}
}

func TestView_WithArchived(t *testing.T) {
	root := setupProject(t)
	archiveDir := filepath.Join(root, "openspec", "archive", "2024-01-01-old-change")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeCommand(t, root, "view")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Archived") {
		t.Errorf("expected 'Archived' in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "old-change") {
		t.Errorf("expected 'old-change' in output, got: %q", stdout)
	}
}

func TestView_WithTasks(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	tasksPath := filepath.Join(root, "openspec", "changes", "add-auth", "tasks.md")
	tasksContent := "- [x] Task 1\n- [ ] Task 2\n"
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeCommand(t, root, "view")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Tasks:") {
		t.Errorf("expected 'Tasks:' in output, got: %q", stdout)
	}
}
