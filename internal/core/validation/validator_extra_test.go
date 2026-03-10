package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/core/schemas"
)

func TestFormatSectionList(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty", nil, ""},
		{"single", []string{"## ADDED Requirements"}, "## ADDED Requirements"},
		{"two", []string{"## ADDED Requirements", "## MODIFIED Requirements"}, "## ADDED Requirements and ## MODIFIED Requirements"},
		{"three", []string{"## ADDED Requirements", "## MODIFIED Requirements", "## REMOVED Requirements"}, "## ADDED Requirements, ## MODIFIED Requirements and ## REMOVED Requirements"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatSectionList(tc.input)
			if got != tc.expected {
				t.Errorf("formatSectionList(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestExtractNameFromPath(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"specs path", "openspec/specs/auth/spec.md", "auth"},
		{"changes path", "openspec/changes/add-auth/proposal.md", "add-auth"},
		{"no recognized segment", "some/random/file.md", "file"},
		{"windows path", `openspec\specs\auth\spec.md`, "auth"},
		{"filename only", "proposal.md", "proposal"},
		{"filename no extension", "README", "README"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractNameFromPath(tc.input)
			if got != tc.expected {
				t.Errorf("extractNameFromPath(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestValidateChangeSchema_EmptyName(t *testing.T) {
	change := &schemas.Change{
		Name:        "",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas:      []schemas.Delta{{Spec: "s", Description: "d"}},
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "name") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about empty name")
	}
}

func TestValidateChangeSchema_WhyTooShort(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         "short",
		WhatChanges: "something",
		Deltas:      []schemas.Delta{{Spec: "s", Description: "d"}},
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "at least") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about why too short")
	}
}

func TestValidateChangeSchema_WhyTooLong(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MaxWhySectionLength+1),
		WhatChanges: "something",
		Deltas:      []schemas.Delta{{Spec: "s", Description: "d"}},
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "exceed") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about why too long")
	}
}

func TestValidateChangeSchema_NoDeltas(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas:      nil,
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "delta") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about no deltas")
	}
}

func TestValidateChangeSchema_TooManyDeltas(t *testing.T) {
	deltas := make([]schemas.Delta, MaxDeltasPerChange+1)
	for i := range deltas {
		deltas[i] = schemas.Delta{Spec: "s", Description: "d"}
	}
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas:      deltas,
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "splitting") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about too many deltas")
	}
}

func TestValidateChangeSchema_DeltaEmptySpec(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas:      []schemas.Delta{{Spec: "", Description: "d"}},
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if issue.Path == "deltas[0].spec" {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about empty delta spec")
	}
}

func TestValidateChangeSchema_DeltaEmptyDescription(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas:      []schemas.Delta{{Spec: "s", Description: ""}},
	}
	issues := validateChangeSchema(change)
	found := false
	for _, issue := range issues {
		if issue.Path == "deltas[0].description" {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about empty delta description")
	}
}

func TestValidateChangeDeltaSpecs_NoSpecsDir(t *testing.T) {
	dir := t.TempDir()

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(dir)

	if report.Valid {
		t.Error("expected invalid report when specs dir missing")
	}
}

func TestValidateChangeDeltaSpecs_ZeroDeltas(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// spec.md with headers but no ### Requirement: blocks
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("## ADDED Requirements\n\nSome text but no requirement blocks\n"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(dir)

	if report.Valid {
		t.Error("expected invalid report for zero deltas")
	}
}

func TestValidateChangeDeltaSpecs_EmptySectionsWithHeaders(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("## ADDED Requirements\n\n## MODIFIED Requirements\n"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(dir)

	if report.Valid {
		t.Error("expected invalid report for empty sections")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "no requirement entries parsed") {
			found = true
		}
	}
	if !found {
		t.Error("expected message about empty sections")
	}
}

func TestValidateChangeDeltaSpecs_CrossSectionConflict_ModifiedAndRemoved(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `## MODIFIED Requirements

### Requirement: User login
The system SHALL authenticate users via password.

#### Scenario: Login
- **WHEN** valid credentials
- **THEN** user is logged in

## REMOVED Requirements

### Requirement: User login
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(dir)

	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "MODIFIED and REMOVED") {
			found = true
		}
	}
	if !found {
		t.Error("expected conflict between MODIFIED and REMOVED")
	}
}

func TestValidateChangeDeltaSpecs_CrossSectionConflict_AddedAndRemoved(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `## ADDED Requirements

### Requirement: User login
The system SHALL authenticate users via password.

#### Scenario: Login
- **WHEN** valid credentials
- **THEN** user is logged in

## REMOVED Requirements

### Requirement: User login
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(dir)

	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "ADDED and REMOVED") {
			found = true
		}
	}
	if !found {
		t.Error("expected conflict between ADDED and REMOVED")
	}
}
