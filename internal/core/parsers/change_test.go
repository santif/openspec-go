package parsers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santif/openspec-go/internal/core/schemas"
)

func TestParseSpecDeltas_AddedOnly(t *testing.T) {
	content := `## ADDED Requirements

### Requirement: User auth
The system SHALL authenticate users.

#### Scenario: Login success
- **WHEN** valid credentials
- **THEN** user is logged in`

	deltas := parseSpecDeltas("my-spec", content)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Operation != schemas.DeltaAdded {
		t.Errorf("Operation = %q, want %q", deltas[0].Operation, schemas.DeltaAdded)
	}
	if deltas[0].Spec != "my-spec" {
		t.Errorf("Spec = %q, want %q", deltas[0].Spec, "my-spec")
	}
}

func TestParseSpecDeltas_ModifiedOnly(t *testing.T) {
	content := `## MODIFIED Requirements

### Requirement: User auth
The system SHALL authenticate users with MFA.

#### Scenario: MFA login
- **WHEN** valid credentials and MFA token
- **THEN** user is logged in`

	deltas := parseSpecDeltas("my-spec", content)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Operation != schemas.DeltaModified {
		t.Errorf("Operation = %q, want %q", deltas[0].Operation, schemas.DeltaModified)
	}
}

func TestParseSpecDeltas_RemovedOnly(t *testing.T) {
	content := `## REMOVED Requirements

### Requirement: Legacy auth
**Reason**: Replaced by new auth
**Migration**: Use new auth endpoint`

	deltas := parseSpecDeltas("my-spec", content)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Operation != schemas.DeltaRemoved {
		t.Errorf("Operation = %q, want %q", deltas[0].Operation, schemas.DeltaRemoved)
	}
}

func TestParseSpecDeltas_RenamedOnly(t *testing.T) {
	content := `## RENAMED Requirements

- FROM: ### Requirement: Old Name
- TO: ### Requirement: New Name`

	deltas := parseSpecDeltas("my-spec", content)
	if len(deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(deltas))
	}
	if deltas[0].Operation != schemas.DeltaRenamed {
		t.Errorf("Operation = %q, want %q", deltas[0].Operation, schemas.DeltaRenamed)
	}
	if deltas[0].Rename == nil {
		t.Fatal("expected Rename to be set")
	}
	if deltas[0].Rename.From != "Old Name" {
		t.Errorf("Rename.From = %q, want %q", deltas[0].Rename.From, "Old Name")
	}
	if deltas[0].Rename.To != "New Name" {
		t.Errorf("Rename.To = %q, want %q", deltas[0].Rename.To, "New Name")
	}
}

func TestParseSpecDeltas_AllSections(t *testing.T) {
	content := `## ADDED Requirements

### Requirement: New feature
The system SHALL do new thing.

#### Scenario: New thing
- **WHEN** triggered
- **THEN** new thing happens

## MODIFIED Requirements

### Requirement: Existing feature
The system SHALL do updated thing.

#### Scenario: Updated thing
- **WHEN** triggered
- **THEN** updated thing happens

## REMOVED Requirements

### Requirement: Old feature
**Reason**: Deprecated
**Migration**: Use new feature

## RENAMED Requirements

- FROM: ### Requirement: Badly Named
- TO: ### Requirement: Well Named`

	deltas := parseSpecDeltas("test", content)
	if len(deltas) != 4 {
		t.Fatalf("expected 4 deltas, got %d", len(deltas))
	}

	ops := map[schemas.DeltaOperation]int{}
	for _, d := range deltas {
		ops[d.Operation]++
	}
	if ops[schemas.DeltaAdded] != 1 {
		t.Errorf("expected 1 ADDED, got %d", ops[schemas.DeltaAdded])
	}
	if ops[schemas.DeltaModified] != 1 {
		t.Errorf("expected 1 MODIFIED, got %d", ops[schemas.DeltaModified])
	}
	if ops[schemas.DeltaRemoved] != 1 {
		t.Errorf("expected 1 REMOVED, got %d", ops[schemas.DeltaRemoved])
	}
	if ops[schemas.DeltaRenamed] != 1 {
		t.Errorf("expected 1 RENAMED, got %d", ops[schemas.DeltaRenamed])
	}
}

func TestParseSpecDeltas_EmptyContent(t *testing.T) {
	deltas := parseSpecDeltas("test", "")
	if len(deltas) != 0 {
		t.Errorf("expected 0 deltas for empty content, got %d", len(deltas))
	}
}

func TestParseRenames_ValidPair(t *testing.T) {
	content := `- FROM: ### Requirement: Old Name
- TO: ### Requirement: New Name`

	renames := parseRenames(content)
	if len(renames) != 1 {
		t.Fatalf("expected 1 rename, got %d", len(renames))
	}
	if renames[0].From != "Old Name" || renames[0].To != "New Name" {
		t.Errorf("got From=%q To=%q", renames[0].From, renames[0].To)
	}
}

func TestParseRenames_WithBackticks(t *testing.T) {
	content := "- FROM: `### Requirement: Old Name`\n- TO: `### Requirement: New Name`"

	renames := parseRenames(content)
	if len(renames) != 1 {
		t.Fatalf("expected 1 rename, got %d", len(renames))
	}
	if renames[0].From != "Old Name" || renames[0].To != "New Name" {
		t.Errorf("got From=%q To=%q", renames[0].From, renames[0].To)
	}
}

func TestParseRenames_IncompletePair_FromOnly(t *testing.T) {
	content := `- FROM: ### Requirement: Old Name`

	renames := parseRenames(content)
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for incomplete pair, got %d", len(renames))
	}
}

func TestParseRenames_MultiplePairs(t *testing.T) {
	content := `- FROM: ### Requirement: First Old
- TO: ### Requirement: First New
- FROM: ### Requirement: Second Old
- TO: ### Requirement: Second New`

	renames := parseRenames(content)
	if len(renames) != 2 {
		t.Fatalf("expected 2 renames, got %d", len(renames))
	}
}

func TestParseRenames_EmptyContent(t *testing.T) {
	renames := parseRenames("")
	if len(renames) != 0 {
		t.Errorf("expected 0 renames for empty content, got %d", len(renames))
	}
}

func TestParseDeltaSpecs_ValidDir(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: Auth
The system SHALL handle auth.

#### Scenario: Login
- **WHEN** user logs in
- **THEN** session is created`), 0644); err != nil {
		t.Fatal(err)
	}

	deltas := parseDeltaSpecs(dir)
	if len(deltas) == 0 {
		t.Fatal("expected deltas")
	}
	if deltas[0].Spec != "auth" {
		t.Errorf("Spec = %q, want %q", deltas[0].Spec, "auth")
	}
}

func TestParseDeltaSpecs_MissingDir(t *testing.T) {
	deltas := parseDeltaSpecs("/nonexistent/path")
	if deltas != nil {
		t.Errorf("expected nil for missing dir, got %v", deltas)
	}
}

func TestParseDeltaSpecs_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	deltas := parseDeltaSpecs(dir)
	if len(deltas) != 0 {
		t.Errorf("expected 0 deltas for empty dir, got %d", len(deltas))
	}
}

func TestParseDeltaSpecs_NonDirEntries(t *testing.T) {
	dir := t.TempDir()
	// Create a file (not a directory) in the specs dir
	if err := os.WriteFile(filepath.Join(dir, "not-a-dir.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	deltas := parseDeltaSpecs(dir)
	if len(deltas) != 0 {
		t.Errorf("expected 0 deltas (only files, no dirs), got %d", len(deltas))
	}
}

func TestParseDeltaSpecs_MissingSpecMd(t *testing.T) {
	dir := t.TempDir()
	// Create a directory but without spec.md
	if err := os.MkdirAll(filepath.Join(dir, "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	deltas := parseDeltaSpecs(dir)
	if len(deltas) != 0 {
		t.Errorf("expected 0 deltas when spec.md missing, got %d", len(deltas))
	}
}

func TestParseChangeWithDeltas_WithDeltaSpecs(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(`## ADDED Requirements

### Requirement: Auth
The system SHALL handle auth.

#### Scenario: Login
- **WHEN** user logs in
- **THEN** session is created`), 0644); err != nil {
		t.Fatal(err)
	}

	content := `## Why

This is needed for security reasons and more text to be valid content here.

## What Changes

- **auth:** Add authentication support`

	change, err := ParseChangeWithDeltas("test-change", content, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if change.Name != "test-change" {
		t.Errorf("Name = %q, want %q", change.Name, "test-change")
	}
	// With delta specs present, should use those instead of simple deltas
	if len(change.Deltas) == 0 {
		t.Fatal("expected deltas from delta specs")
	}
	if change.Deltas[0].Spec != "auth" {
		t.Errorf("Deltas[0].Spec = %q, want %q", change.Deltas[0].Spec, "auth")
	}
}

func TestParseChangeWithDeltas_NoSpecsDir(t *testing.T) {
	dir := t.TempDir()
	content := `## Why

This is needed for good reasons and more text to meet the minimum length.

## What Changes

- **auth:** Add authentication support`

	change, err := ParseChangeWithDeltas("test", content, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Falls back to simple deltas from What Changes
	if len(change.Deltas) == 0 {
		t.Error("expected fallback to simple deltas")
	}
}

func TestParseChangeWithDeltas_MissingWhy(t *testing.T) {
	dir := t.TempDir()
	content := `## What Changes

- **auth:** Add authentication support`

	_, err := ParseChangeWithDeltas("test", content, dir)
	if err == nil {
		t.Fatal("expected error for missing Why section")
	}
}

func TestParseChangeWithDeltas_MissingWhatChanges(t *testing.T) {
	dir := t.TempDir()
	content := `## Why

This is needed for good reasons and more text.`

	_, err := ParseChangeWithDeltas("test", content, dir)
	if err == nil {
		t.Fatal("expected error for missing What Changes section")
	}
}
