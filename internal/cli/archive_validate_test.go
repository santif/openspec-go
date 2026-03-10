package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// validProposal is a proposal that passes all validation checks, including
// the minimum length requirement for the Why section (50 characters).
var validProposal = `# add-auth

## Why

We need to add authentication to our application to ensure that only authorized users can access protected resources and sensitive data.

## What Changes

- **auth-service:** Add authentication service

## Impact

Backend API changes.
`

// sampleDeltaSpec is a valid delta spec with an ADDED requirement that includes
// a normative keyword and a scenario (required by the validator).
var sampleDeltaSpec = `# auth-service

## ADDED Requirements

### Requirement: Login

Users SHALL be able to login with email and password.

#### Scenario: Successful login

- **WHEN** valid credentials provided
- **THEN** user receives a session token
`

// writeChangeDeltaSpec writes a delta spec file inside a change's specs directory.
func writeChangeDeltaSpec(t *testing.T, root, changeName, specName, content string) {
	t.Helper()
	dir := filepath.Join(root, "openspec", "changes", changeName, "specs", specName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// --- archive command ---

func TestArchive_NoChanges(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "archive", "--yes")
	if err == nil {
		t.Fatal("expected error when no changes exist")
	}
	if !strings.Contains(err.Error(), "no active changes") {
		t.Errorf("expected error containing 'no active changes', got: %v", err)
	}
}

func TestArchive_ChangeNotFound(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "archive", "nonexistent", "--yes")
	if err == nil {
		t.Fatal("expected error for nonexistent change")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error containing 'not found', got: %v", err)
	}
}

func TestArchive_SingleChange_SkipSpecs(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	_, _, err := executeCommand(t, root, "archive", "--yes", "--skip-specs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Change directory should no longer exist
	changeDir := filepath.Join(root, "openspec", "changes", "add-auth")
	if _, statErr := os.Stat(changeDir); !os.IsNotExist(statErr) {
		t.Error("expected change directory to be removed after archive")
	}

	// Archive directory should contain a directory with "add-auth" in its name
	archiveDir := filepath.Join(root, "openspec", "archive")
	entries, readErr := os.ReadDir(archiveDir)
	if readErr != nil {
		t.Fatalf("failed to read archive directory: %v", readErr)
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "add-auth") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected archive directory to contain an entry with 'add-auth' in its name")
	}
}

func TestArchive_NamedChange_NoValidate(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	_, _, err := executeCommand(t, root, "archive", "add-auth", "--yes", "--no-validate", "--skip-specs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the change was archived
	changeDir := filepath.Join(root, "openspec", "changes", "add-auth")
	if _, statErr := os.Stat(changeDir); !os.IsNotExist(statErr) {
		t.Error("expected change directory to be removed after archive")
	}
}

func TestArchive_MultipleChanges_NoArg(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)
	writeChange(t, root, "add-logging", sampleProposal)

	_, _, err := executeCommand(t, root, "archive", "--yes")
	if err == nil {
		t.Fatal("expected error when multiple changes exist without specifying one")
	}
	if !strings.Contains(err.Error(), "multiple changes found") {
		t.Errorf("expected error containing 'multiple changes found', got: %v", err)
	}
}

func TestArchive_ValidationFailure(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", "bad content no sections")

	// Without --yes, validation failure returns an error immediately
	_, _, err := executeCommand(t, root, "archive", "add-auth")
	if err == nil {
		t.Fatal("expected error for invalid proposal content")
	}
}

// --- validate text output (printValidationResult) ---

func TestValidate_TextOutput_Valid(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", validProposal)
	writeChangeDeltaSpec(t, root, "add-auth", "auth-service", sampleDeltaSpec)

	stdout, _, err := executeCommand(t, root, "validate", "add-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "OK") {
		t.Errorf("expected 'OK' in text output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "add-auth") {
		t.Errorf("expected 'add-auth' in text output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Summary:") {
		t.Errorf("expected 'Summary:' in text output, got: %q", stdout)
	}
}

func TestValidate_TextOutput_Spec(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "validate", "user-auth", "--type", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "OK") {
		t.Errorf("expected 'OK' in text output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "user-auth") {
		t.Errorf("expected 'user-auth' in text output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Summary:") {
		t.Errorf("expected 'Summary:' in text output, got: %q", stdout)
	}
}

func TestValidate_TextOutput_AllText(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", validProposal)
	writeChangeDeltaSpec(t, root, "add-auth", "auth-service", sampleDeltaSpec)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "validate", "--all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain two "Summary:" lines, one per result
	count := strings.Count(stdout, "Summary:")
	if count != 2 {
		t.Errorf("expected 2 'Summary:' lines in text output, got %d; output: %q", count, stdout)
	}
}
