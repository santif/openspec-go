package converters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santif/openspec-go/internal/core/schemas"
)

func TestConvertSpecToJSON(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "openspec", "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	specContent := `# Auth Specification

## Purpose
Handle user authentication.

## Requirements

### Requirement: Login Flow
Users SHALL be able to log in with email and password.

#### Scenario: Successful login
Given valid credentials, the user is authenticated.
`
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertSpecToJSON(specPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec schemas.Spec
	if err := json.Unmarshal([]byte(result), &spec); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if spec.Name != "auth" {
		t.Errorf("expected name 'auth', got %q", spec.Name)
	}
	if spec.Metadata == nil || spec.Metadata.SourcePath != specPath {
		t.Error("expected metadata.sourcePath to be set")
	}
	if len(spec.Requirements) == 0 {
		t.Error("expected at least one requirement")
	}
}

func TestConvertChangeToJSON(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "openspec", "changes", "add-auth")
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		t.Fatal(err)
	}

	changeContent := `# Add Auth

## Why
We need user authentication.

## What Changes
- **auth:** Add login and registration flows
`
	changePath := filepath.Join(changeDir, "proposal.md")
	if err := os.WriteFile(changePath, []byte(changeContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertChangeToJSON(changePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var change schemas.Change
	if err := json.Unmarshal([]byte(result), &change); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if change.Name != "add-auth" {
		t.Errorf("expected name 'add-auth', got %q", change.Name)
	}
	if change.Metadata == nil || change.Metadata.SourcePath != changePath {
		t.Error("expected metadata.sourcePath to be set")
	}
}

func TestConvertSpecToJSON_FileNotFound(t *testing.T) {
	_, err := ConvertSpecToJSON("/nonexistent/path/spec.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestConvertChangeToJSON_FileNotFound(t *testing.T) {
	_, err := ConvertChangeToJSON("/nonexistent/path/proposal.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestExtractNameFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"openspec/specs/auth/spec.md", "auth"},
		{"openspec/changes/add-feature/proposal.md", "add-feature"},
		{"/abs/path/openspec/specs/billing/spec.md", "billing"},
		{"some/random/file.md", "file"},
		{"noext", "noext"},
	}

	for _, tt := range tests {
		got := extractNameFromPath(tt.path)
		if got != tt.want {
			t.Errorf("extractNameFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestConvertSpecToJSON_ParseError(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "openspec", "specs", "bad")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte("No valid sections here"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ConvertSpecToJSON(specPath)
	if err == nil {
		t.Error("expected error for unparseable spec")
	}
}

func TestConvertChangeToJSON_ParseError(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "openspec", "changes", "bad")
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		t.Fatal(err)
	}

	changePath := filepath.Join(changeDir, "proposal.md")
	if err := os.WriteFile(changePath, []byte("No valid sections"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ConvertChangeToJSON(changePath)
	if err == nil {
		t.Error("expected error for unparseable change")
	}
}

func TestConvertSpecToJSON_MetadataPopulated(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "openspec", "specs", "billing")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	specContent := `## Purpose
Handle billing and payments for all users.

## Requirements

### Requirement: Payment Processing
The system SHALL process payments securely.

#### Scenario: Successful payment
Given a valid credit card, the payment is processed.
`
	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertSpecToJSON(specPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var spec schemas.Spec
	if err := json.Unmarshal([]byte(result), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if spec.Name != "billing" {
		t.Errorf("expected name 'billing', got %q", spec.Name)
	}
	if spec.Metadata == nil {
		t.Fatal("expected metadata to be populated")
	}
	if spec.Metadata.SourcePath != specPath {
		t.Errorf("expected sourcePath %q, got %q", specPath, spec.Metadata.SourcePath)
	}
}

func TestConvertChangeToJSON_WithDeltas(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "openspec", "changes", "add-billing")
	specsDir := filepath.Join(changeDir, "specs", "billing")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	changeContent := `## Why
We need billing capabilities for the platform.

## What Changes
- **billing:** Add billing and payment processing capabilities
`
	changePath := filepath.Join(changeDir, "proposal.md")
	if err := os.WriteFile(changePath, []byte(changeContent), 0644); err != nil {
		t.Fatal(err)
	}

	deltaContent := `## ADDED Requirements

### Requirement: Payment
The system SHALL process payments.

#### Scenario: Payment success
Given valid payment info, process the payment.
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ConvertChangeToJSON(changePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var change schemas.Change
	if err := json.Unmarshal([]byte(result), &change); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if change.Name != "add-billing" {
		t.Errorf("expected name 'add-billing', got %q", change.Name)
	}
	if len(change.Deltas) == 0 {
		t.Error("expected at least one delta")
	}
}
