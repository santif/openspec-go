package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSpec_Valid(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	content := `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: User Login
The system SHALL allow users to login with valid credentials.

#### Scenario: Successful login
- **WHEN** user enters valid username and password
- **THEN** system grants access and creates a session
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateSpec(specFile)

	if !report.Valid {
		t.Errorf("expected valid report, got issues: %v", report.Issues)
	}
}

func TestValidateSpec_MissingPurpose(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	content := `## Requirements

### Requirement: Feature
The system SHALL do something.

#### Scenario: T
- **WHEN** test
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateSpec(specFile)

	if report.Valid {
		t.Error("expected invalid report for missing Purpose")
	}
}

func TestValidateSpec_MissingRequirements(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	content := `## Purpose
A comprehensive authentication and authorization system that manages user access.
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateSpec(specFile)

	if report.Valid {
		t.Error("expected invalid report for missing Requirements")
	}
}

func TestValidateSpec_NoShallOrMust(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Feature
The system does something without required keywords.

#### Scenario: T
- **WHEN** test
- **THEN** result
`)

	if report.Valid {
		t.Error("expected invalid report for missing SHALL/MUST")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "SHALL or MUST") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about missing SHALL or MUST keyword")
	}
}

func TestValidateSpec_NoScenarios(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Feature
The system SHALL do something.
`)

	if report.Valid {
		t.Error("expected invalid report for missing scenarios")
	}
}

func TestValidateSpecContent(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateSpecContent("test-spec", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid credentials
- **WHEN** user provides valid credentials
- **THEN** system authenticates them
`)

	if !report.Valid {
		t.Errorf("expected valid report, got issues: %v", report.Issues)
	}
}

func TestValidateChange_Valid(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "add-auth")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	content := `## Why
This change adds authentication to the system which is critical for security and user management.

## What Changes
- **auth:** Adds new authentication requirements
`
	if err := os.WriteFile(proposalFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a delta spec so deltas are found
	deltaContent := `## ADDED Requirements

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid login
- **WHEN** user logs in
- **THEN** system grants access
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if !report.Valid {
		t.Errorf("expected valid report, got issues: %v", report.Issues)
	}
}

func TestValidateChange_MissingWhy(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	content := `## What Changes
- **auth:** Adds something
`
	if err := os.WriteFile(proposalFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if report.Valid {
		t.Error("expected invalid report for missing Why")
	}
}

func TestValidateChange_WhyTooShort(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	content := `## Why
Short.

## What Changes
- **auth:** Adds authentication
`
	if err := os.WriteFile(proposalFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	deltaContent := "## ADDED Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if report.Valid {
		t.Error("expected invalid report for short Why")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "at least") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about Why being too short")
	}
}

func TestValidateChange_WhyTooLong(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	longWhy := strings.Repeat("This is a very long explanation. ", 100)
	content := "## Why\n" + longWhy + "\n\n## What Changes\n- **auth:** Adds authentication\n"
	if err := os.WriteFile(proposalFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	deltaContent := "## ADDED Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if report.Valid {
		t.Error("expected invalid report for long Why")
	}
}

func TestValidateChange_TooManyDeltas(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	var deltas strings.Builder
	for i := 0; i < 12; i++ {
		deltas.WriteString("- **spec-" + string(rune('a'+i)) + ":** Adds requirements\n")
	}
	content := "## Why\nThis is a detailed explanation of why this change is needed with enough characters.\n\n## What Changes\n" + deltas.String()
	if err := os.WriteFile(proposalFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if report.Valid {
		t.Error("expected invalid report for too many deltas")
	}
}

func TestValidateChangeDeltaSpecs_DuplicateInSection(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## ADDED Requirements

### Requirement: Login
The system SHALL allow login.

#### Scenario: T
- **WHEN** test

### Requirement: Login
The system SHALL allow login again.

#### Scenario: T2
- **WHEN** test2
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for duplicate in ADDED")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "Duplicate") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about duplicate requirement")
	}
}

func TestValidateChangeDeltaSpecs_CrossSectionConflict(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## MODIFIED Requirements

### Requirement: Login
The system SHALL modify login.

#### Scenario: T
- **WHEN** test

## REMOVED Requirements

### Requirement: Login
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for cross-section conflict")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "MODIFIED") && strings.Contains(issue.Message, "REMOVED") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected issue about cross-section conflict")
	}
}

func TestValidateChangeDeltaSpecs_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## added requirements

### Requirement: Login
The system SHALL allow login.

#### Scenario: T
- **WHEN** test
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if !report.Valid {
		t.Errorf("expected valid report, got issues: %v", report.Issues)
	}
}

func TestStrictMode_WarningsAsErrors(t *testing.T) {
	v := NewValidator(true)
	// This spec has a short purpose which triggers a warning
	report := v.ValidateSpecContent("test", `## Purpose
Short.

## Requirements

### Requirement: Feature
The system SHALL do something.

#### Scenario: T
- **WHEN** test
- **THEN** result
`)

	if report.Valid {
		t.Error("expected strict mode to treat warnings as errors")
	}
	if report.Summary.Warnings == 0 {
		t.Error("expected at least one warning")
	}
}

func TestNonStrictMode_WarningsAllowed(t *testing.T) {
	v := NewValidator(false)
	// This spec has a short purpose which triggers a warning
	report := v.ValidateSpecContent("test", `## Purpose
Short.

## Requirements

### Requirement: Feature
The system SHALL do something.

#### Scenario: T
- **WHEN** test
- **THEN** result
`)

	if !report.Valid {
		// In non-strict mode, warnings should not make the report invalid
		hasOnlyWarnings := true
		for _, issue := range report.Issues {
			if issue.Level == LevelError {
				hasOnlyWarnings = false
				break
			}
		}
		if hasOnlyWarnings {
			t.Error("expected non-strict mode to allow warnings")
		}
	}
}
