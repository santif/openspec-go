package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/core/schemas"
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

func TestNewValidatorWithKeywords_CustomKeywords(t *testing.T) {
	v := NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, nil)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Login
El sistema DEBE permitir el acceso con credenciales válidas.

#### Scenario: Successful login
- **WHEN** user enters valid username and password
- **THEN** system grants access
`)

	if !report.Valid {
		t.Errorf("expected valid report with custom keyword DEBE, got issues: %v", report.Issues)
	}
}

func TestNewValidatorWithKeywords_CustomKeywordsFail(t *testing.T) {
	v := NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, nil)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Login
The system allows login with valid credentials.

#### Scenario: Successful login
- **WHEN** user enters valid username and password
- **THEN** system grants access
`)

	if report.Valid {
		t.Error("expected invalid report when custom keywords are missing")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "DEBE") && strings.Contains(issue.Message, "DEBERA") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error message to mention DEBE and DEBERA")
	}
}

func TestNewValidatorWithKeywords_NilFallsBackToDefaults(t *testing.T) {
	v := NewValidatorWithKeywords(false, nil, nil)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid credentials
- **WHEN** user provides valid credentials
- **THEN** system authenticates them
`)

	if !report.Valid {
		t.Errorf("expected valid report with default keywords, got issues: %v", report.Issues)
	}
}

func TestNewValidatorWithKeywords_AccentedCharacters(t *testing.T) {
	v := NewValidatorWithKeywords(false, []string{"DEBERÁ", "DEBE"}, nil)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Login
El sistema DEBERÁ permitir el acceso.

#### Scenario: Successful login
- **WHEN** user enters valid credentials
- **THEN** system grants access
`)

	if !report.Valid {
		t.Errorf("expected valid report with accented keyword DEBERÁ, got issues: %v", report.Issues)
	}
}

func TestNewValidatorWithKeywords_DeltaSpecValidation(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## ADDED Requirements

### Requirement: Login
El sistema DEBE permitir login.

#### Scenario: T
- **WHEN** test
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, nil)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if !report.Valid {
		t.Errorf("expected valid report with custom keywords in delta spec, got issues: %v", report.Issues)
	}
}

func TestNewValidatorWithKeywords_DeltaSpecFails(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## ADDED Requirements

### Requirement: Login
The system allows login.

#### Scenario: T
- **WHEN** test
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, nil)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report when custom keywords missing from delta spec")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "DEBE") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error message to mention custom keywords")
	}
}

func TestNewValidatorWithKeywords_ErrorMessageDefaultKeywords(t *testing.T) {
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
		t.Error("expected invalid report")
	}
	found := false
	for _, issue := range report.Issues {
		if issue.Message == "Requirement must contain SHALL or MUST keyword" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected default keyword error message, got issues: %v", report.Issues)
	}
}

func TestNewValidatorWithKeywords_ErrorMessageCustomKeywords(t *testing.T) {
	v := NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, nil)
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
		t.Error("expected invalid report")
	}
	found := false
	for _, issue := range report.Issues {
		if issue.Message == "Requirement must contain DEBE or DEBERA keyword" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected custom keyword error message, got issues: %v", report.Issues)
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

func TestValidateSpec_FileNotFound(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateSpec("/nonexistent/path/spec.md")

	if report.Valid {
		t.Error("expected invalid report for non-existent file")
	}
	if report.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", report.Summary.Errors)
	}
}

func TestValidateSpec_ParseError(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", "bad")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	// Content with no sections at all triggers a parse error
	if err := os.WriteFile(specFile, []byte("Just some text with no sections"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateSpec(specFile)

	if report.Valid {
		t.Error("expected invalid report for unparseable spec")
	}
}

func TestValidateSpecContent_ParseError(t *testing.T) {
	v := NewValidator(false)
	// No valid sections
	report := v.ValidateSpecContent("bad", "No sections here")

	if report.Valid {
		t.Error("expected invalid report for unparseable content")
	}
	if report.Summary.Errors == 0 {
		t.Error("expected at least one error")
	}
}

func TestValidateChange_FileNotFound(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateChange("/nonexistent/changes/test/proposal.md")

	if report.Valid {
		t.Error("expected invalid report for non-existent file")
	}
	if report.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", report.Summary.Errors)
	}
}

func TestValidateChange_ParseError(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "bad")
	if err := os.MkdirAll(changeDir, 0755); err != nil {
		t.Fatal(err)
	}
	proposalFile := filepath.Join(changeDir, "proposal.md")
	if err := os.WriteFile(proposalFile, []byte("No sections at all"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChange(proposalFile)

	if report.Valid {
		t.Error("expected invalid report for unparseable change")
	}
}

func TestValidateChangeDeltaSpecs_EmptySection(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Section header present but no requirement blocks
	content := "## ADDED Requirements\n\nSome text but no requirement blocks.\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for empty delta section")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "no requirement entries parsed") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about empty sections, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_MissingHeaders(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No delta section headers at all
	content := "Some random content without delta headers\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for missing delta headers")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "No delta sections found") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing delta headers, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_MissingRequirementText(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Requirement header with no text body
	content := `## ADDED Requirements

### Requirement: EmptyReq

#### Scenario: T
- **WHEN** test
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for missing requirement text")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "missing requirement text") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing requirement text, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_MissingScenarios(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Requirement with text but no scenarios
	content := `## ADDED Requirements

### Requirement: NoScenario
The system SHALL do something important.
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for missing scenarios")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "at least one scenario") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing scenarios, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_DuplicateRemoved(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## REMOVED Requirements

### Requirement: Login

### Requirement: Login
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for duplicate in REMOVED")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "Duplicate") && strings.Contains(issue.Message, "REMOVED") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate error in REMOVED, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_DuplicateModified(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `## MODIFIED Requirements

### Requirement: Login
The system SHALL allow login.

#### Scenario: T
- **WHEN** test

### Requirement: Login
The system SHALL allow another login.

#### Scenario: T2
- **WHEN** test2
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for duplicate in MODIFIED")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "Duplicate") && strings.Contains(issue.Message, "MODIFIED") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate error in MODIFIED, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_AddedAndRemoved(t *testing.T) {
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

## REMOVED Requirements

### Requirement: Login
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for ADDED and REMOVED conflict")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "ADDED") && strings.Contains(issue.Message, "REMOVED") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cross-section conflict, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_ModifiedAndAdded(t *testing.T) {
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

## MODIFIED Requirements

### Requirement: Login
The system SHALL modify login.

#### Scenario: T2
- **WHEN** test2
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for MODIFIED and ADDED conflict")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "MODIFIED") && strings.Contains(issue.Message, "ADDED") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cross-section conflict, got: %v", report.Issues)
	}
}

func TestValidateSpecSchema_EmptyScenarioText(t *testing.T) {
	v := NewValidator(false)
	report := v.ValidateSpecContent("test", `## Purpose
A comprehensive authentication and authorization system that manages user access.

## Requirements

### Requirement: Feature
The system SHALL do something.

#### Scenario:
`)

	if report.Valid {
		t.Error("expected invalid report for empty scenario text")
	}
}

func TestApplySpecRules_RequirementTooLong(t *testing.T) {
	v := NewValidator(false)
	longText := strings.Repeat("The system SHALL perform validation. ", 20)
	report := v.ValidateSpecContent("test", "## Purpose\nA comprehensive authentication and authorization system that manages user access.\n\n## Requirements\n\n### Requirement: Feature\n"+longText+"\n\n#### Scenario: T\n- **WHEN** test\n")

	foundInfo := false
	for _, issue := range report.Issues {
		if issue.Level == LevelInfo && strings.Contains(issue.Message, "long") {
			foundInfo = true
			break
		}
	}
	if !foundInfo {
		t.Errorf("expected info about long requirement, got: %v", report.Issues)
	}
}

func TestApplyChangeRules_BriefDeltaDescription(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas: []schemas.Delta{
			{Spec: "auth", Description: "Short", Operation: schemas.DeltaAdded},
		},
	}
	issues := applyChangeRules(change)
	foundBrief := false
	for _, issue := range issues {
		if issue.Level == LevelWarning && strings.Contains(issue.Message, "brief") {
			foundBrief = true
			break
		}
	}
	if !foundBrief {
		t.Errorf("expected warning about brief delta description, got: %v", issues)
	}
}

func TestApplyChangeRules_MissingRequirements(t *testing.T) {
	change := &schemas.Change{
		Name:        "test",
		Why:         strings.Repeat("x", MinWhySectionLength),
		WhatChanges: "something",
		Deltas: []schemas.Delta{
			{Spec: "auth", Description: "Add authentication flow to system", Operation: schemas.DeltaAdded, Requirements: nil},
		},
	}
	issues := applyChangeRules(change)
	foundMissing := false
	for _, issue := range issues {
		if issue.Level == LevelWarning && strings.Contains(issue.Message, "requirements") {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Errorf("expected warning about missing requirements, got: %v", issues)
	}
}

func TestEnrichTopLevelError_MissingSpecSections(t *testing.T) {
	v := NewValidator(false)
	msg := v.enrichTopLevelError("spec must have a Purpose section")
	if !strings.Contains(msg, "**WHEN**") {
		t.Error("expected guide with WHEN keyword for missing spec sections")
	}

	msg = v.enrichTopLevelError("spec must have a Requirements section")
	if !strings.Contains(msg, "**WHEN**") {
		t.Error("expected guide with WHEN keyword for missing requirements section")
	}
}

func TestEnrichTopLevelError_MissingChangeSections(t *testing.T) {
	v := NewValidator(false)
	msg := v.enrichTopLevelError("change must have a Why section")
	if !strings.Contains(msg, Messages.GuideMissingChangeSections) {
		t.Error("expected guide for missing change sections")
	}

	msg = v.enrichTopLevelError("change must have a What Changes section")
	if !strings.Contains(msg, Messages.GuideMissingChangeSections) {
		t.Error("expected guide for missing change sections")
	}
}

func TestEnrichTopLevelError_NoDeltas(t *testing.T) {
	v := NewValidator(false)
	msg := v.enrichTopLevelError(Messages.ChangeNoDeltas)
	if !strings.Contains(msg, Messages.GuideNoDeltas) {
		t.Error("expected guide for no deltas")
	}
}

func TestEnrichTopLevelError_UnknownMessage(t *testing.T) {
	v := NewValidator(false)
	msg := v.enrichTopLevelError("some unknown error message")
	if msg != "some unknown error message" {
		t.Errorf("expected passthrough for unknown message, got: %q", msg)
	}
}

func TestGuideMessages_DefaultConditionals(t *testing.T) {
	v := NewValidator(false)
	guide := v.guideScenarioFormat()
	if !strings.Contains(guide, "**WHEN**") || !strings.Contains(guide, "**THEN**") || !strings.Contains(guide, "**AND**") {
		t.Errorf("expected default WHEN/THEN/AND in guide, got: %s", guide)
	}
	missing := v.guideMissingSpecSections()
	if !strings.Contains(missing, "**WHEN**") || !strings.Contains(missing, "**THEN**") {
		t.Errorf("expected default WHEN/THEN in missing sections guide, got: %s", missing)
	}
}

func TestGuideMessages_CustomConditionals(t *testing.T) {
	cond := &projectconfig.ConditionalsConfig{When: "CUANDO", Then: "ENTONCES", And: "Y"}
	v := NewValidatorWithKeywords(false, nil, cond)
	guide := v.guideScenarioFormat()
	if !strings.Contains(guide, "**CUANDO**") || !strings.Contains(guide, "**ENTONCES**") || !strings.Contains(guide, "**Y**") {
		t.Errorf("expected CUANDO/ENTONCES/Y in guide, got: %s", guide)
	}
	if strings.Contains(guide, "**WHEN**") || strings.Contains(guide, "**THEN**") {
		t.Errorf("should not contain default WHEN/THEN when custom conditionals set, got: %s", guide)
	}
	missing := v.guideMissingSpecSections()
	if !strings.Contains(missing, "**CUANDO**") || !strings.Contains(missing, "**ENTONCES**") {
		t.Errorf("expected CUANDO/ENTONCES in missing sections guide, got: %s", missing)
	}
}

func TestExtractRequirementText_WithMetadata(t *testing.T) {
	block := "### Requirement: Feature\n**Priority**: High\nThe system SHALL do something.\n\n#### Scenario: T\n"
	text := extractRequirementText(block)
	if text != "The system SHALL do something." {
		t.Errorf("expected requirement text, got %q", text)
	}
}

func TestExtractRequirementText_EmptyBlock(t *testing.T) {
	block := "### Requirement: Feature\n"
	text := extractRequirementText(block)
	if text != "" {
		t.Errorf("expected empty text, got %q", text)
	}
}

func TestExtractRequirementText_OnlyScenarios(t *testing.T) {
	block := "### Requirement: Feature\n#### Scenario: T\n- **WHEN** test\n"
	text := extractRequirementText(block)
	if text != "" {
		t.Errorf("expected empty text when only scenarios present, got %q", text)
	}
}

func TestCountScenarios(t *testing.T) {
	tests := []struct {
		name  string
		block string
		want  int
	}{
		{"no scenarios", "Some text\nMore text\n", 0},
		{"one scenario", "Text\n#### Scenario: T\n- test\n", 1},
		{"two scenarios", "Text\n#### Scenario: A\n- a\n#### Scenario: B\n- b\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countScenarios(tt.block)
			if got != tt.want {
				t.Errorf("countScenarios = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCreateReport_SummaryCounts(t *testing.T) {
	v := NewValidator(false)
	issues := []Issue{
		{Level: LevelError, Path: "a", Message: "err1"},
		{Level: LevelError, Path: "b", Message: "err2"},
		{Level: LevelWarning, Path: "c", Message: "warn1"},
		{Level: LevelInfo, Path: "d", Message: "info1"},
	}

	report := v.createReport(issues)

	if report.Summary.Errors != 2 {
		t.Errorf("expected 2 errors, got %d", report.Summary.Errors)
	}
	if report.Summary.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", report.Summary.Warnings)
	}
	if report.Summary.Info != 1 {
		t.Errorf("expected 1 info, got %d", report.Summary.Info)
	}
	if report.Valid {
		t.Error("expected invalid report with errors")
	}
}

func TestCreateReport_ValidWithNoIssues(t *testing.T) {
	v := NewValidator(false)
	report := v.createReport(nil)

	if !report.Valid {
		t.Error("expected valid report with no issues")
	}
	if report.Summary.Errors != 0 || report.Summary.Warnings != 0 || report.Summary.Info != 0 {
		t.Error("expected all zero counts")
	}
}

func TestCreateReport_StrictWithWarnings(t *testing.T) {
	v := NewValidator(true)
	issues := []Issue{
		{Level: LevelWarning, Path: "a", Message: "warn"},
	}

	report := v.createReport(issues)

	if report.Valid {
		t.Error("expected strict mode to treat warnings as invalid")
	}
}

func TestValidateSpecSchema_EmptyRequirementText(t *testing.T) {
	v := NewValidator(false)
	spec := &schemas.Spec{
		Name:     "test",
		Overview: strings.Repeat("x", MinPurposeLength+1),
		Requirements: []schemas.Requirement{
			{Text: "", Scenarios: []schemas.Scenario{{RawText: "some scenario"}}},
		},
	}
	issues := v.validateSpecSchema(spec)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "empty") && issue.Path == "requirements[0].text" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about empty requirement text, got: %v", issues)
	}
}

func TestValidateSpecSchema_EmptyScenarioRawText(t *testing.T) {
	v := NewValidator(false)
	spec := &schemas.Spec{
		Name:     "test",
		Overview: strings.Repeat("x", MinPurposeLength+1),
		Requirements: []schemas.Requirement{
			{Text: "The system SHALL do something", Scenarios: []schemas.Scenario{{RawText: ""}}},
		},
	}
	issues := v.validateSpecSchema(spec)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Scenario") && strings.Contains(issue.Message, "empty") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about empty scenario, got: %v", issues)
	}
}

func TestValidateSpecSchema_EmptyNameAndOverview(t *testing.T) {
	v := NewValidator(false)
	spec := &schemas.Spec{
		Name:         "",
		Overview:     "",
		Requirements: nil,
	}
	issues := v.validateSpecSchema(spec)

	var foundName, foundOverview, foundReqs bool
	for _, issue := range issues {
		if strings.Contains(issue.Message, "name") {
			foundName = true
		}
		if strings.Contains(issue.Message, "Purpose") {
			foundOverview = true
		}
		if strings.Contains(issue.Message, "requirement") {
			foundReqs = true
		}
	}
	if !foundName {
		t.Error("expected error about empty name")
	}
	if !foundOverview {
		t.Error("expected error about empty overview/purpose")
	}
	if !foundReqs {
		t.Error("expected error about no requirements")
	}
}

func TestValidateChangeDeltaSpecs_RenamedDuplicateFrom(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "## RENAMED Requirements\n\n" +
		"- FROM: `### Requirement: Old Name`\n" +
		"- TO: `### Requirement: New Name A`\n\n" +
		"- FROM: `### Requirement: Old Name`\n" +
		"- TO: `### Requirement: New Name B`\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for duplicate FROM in RENAMED")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "Duplicate FROM") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about duplicate FROM in RENAMED, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_RenamedDuplicateTo(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	specsDir := filepath.Join(changeDir, "specs", "auth")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "## RENAMED Requirements\n\n" +
		"- FROM: `### Requirement: Old A`\n" +
		"- TO: `### Requirement: New Name`\n\n" +
		"- FROM: `### Requirement: Old B`\n" +
		"- TO: `### Requirement: New Name`\n"
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if report.Valid {
		t.Error("expected invalid report for duplicate TO in RENAMED")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "Duplicate TO") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about duplicate TO in RENAMED, got: %v", report.Issues)
	}
}

func TestValidateChangeDeltaSpecs_MultipleSpecs(t *testing.T) {
	dir := t.TempDir()
	changeDir := filepath.Join(dir, "changes", "test")
	authSpecsDir := filepath.Join(changeDir, "specs", "auth")
	billingSpecsDir := filepath.Join(changeDir, "specs", "billing")
	if err := os.MkdirAll(authSpecsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(billingSpecsDir, 0755); err != nil {
		t.Fatal(err)
	}

	authContent := `## ADDED Requirements

### Requirement: Login
The system SHALL authenticate users.

#### Scenario: Valid login
- **WHEN** user logs in
`
	billingContent := `## ADDED Requirements

### Requirement: Payment
The system SHALL process payments.

#### Scenario: Valid payment
- **WHEN** user pays
`
	if err := os.WriteFile(filepath.Join(authSpecsDir, "spec.md"), []byte(authContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(billingSpecsDir, "spec.md"), []byte(billingContent), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(false)
	report := v.ValidateChangeDeltaSpecs(changeDir)

	if !report.Valid {
		t.Errorf("expected valid report for multiple valid specs, got: %v", report.Issues)
	}
}
