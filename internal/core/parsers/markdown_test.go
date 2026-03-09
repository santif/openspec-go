package parsers

import (
	"strings"
	"testing"
)

func TestNormalizeContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"CRLF", "line1\r\nline2\r\n", "line1\nline2\n"},
		{"CR only", "line1\rline2\r", "line1\nline2\n"},
		{"LF only", "line1\nline2\n", "line1\nline2\n"},
		{"Mixed", "a\r\nb\rc\n", "a\nb\nc\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeContent(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestParseSections_NestedHeaders(t *testing.T) {
	content := `## Top
content

### Child
child content

#### Grandchild
grandchild content
`

	sections := ParseSections(content)

	if len(sections) != 1 {
		t.Fatalf("expected 1 top-level section, got %d", len(sections))
	}
	if sections[0].Title != "Top" {
		t.Errorf("expected title 'Top', got %q", sections[0].Title)
	}
	if len(sections[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(sections[0].Children))
	}
	if sections[0].Children[0].Title != "Child" {
		t.Errorf("expected child title 'Child', got %q", sections[0].Children[0].Title)
	}
	if len(sections[0].Children[0].Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(sections[0].Children[0].Children))
	}
}

func TestParseSections_ContentBetweenHeaders(t *testing.T) {
	content := `## Section One
Line 1 of section one.
Line 2 of section one.

## Section Two
Content of section two.
`

	sections := ParseSections(content)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if !strings.Contains(sections[0].Content, "Line 1") {
		t.Error("expected section one to contain 'Line 1'")
	}
	if !strings.Contains(sections[0].Content, "Line 2") {
		t.Error("expected section one to contain 'Line 2'")
	}
}

func TestFindSection_CaseInsensitive(t *testing.T) {
	content := `## PURPOSE
Some purpose.

## REQUIREMENTS
Some requirements.
`
	sections := ParseSections(content)
	found := FindSection(sections, "purpose")

	if found == nil {
		t.Fatal("expected to find section 'purpose' case-insensitively")
	}
	if found.Title != "PURPOSE" {
		t.Errorf("expected title 'PURPOSE', got %q", found.Title)
	}
}

func TestFindSection_NotFound(t *testing.T) {
	content := `## Purpose
Some purpose.
`
	sections := ParseSections(content)
	found := FindSection(sections, "NonExistent")

	if found != nil {
		t.Error("expected nil for non-existent section")
	}
}

func TestParseSpec_Valid(t *testing.T) {
	content := `## Purpose
This is a valid purpose for the specification.

## Requirements

### Requirement: Feature A
The system SHALL do feature A.

#### Scenario: Basic use
- **WHEN** user does X
- **THEN** system does Y
`

	spec, err := ParseSpec("test-spec", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "test-spec" {
		t.Errorf("expected name 'test-spec', got %q", spec.Name)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(spec.Requirements))
	}
}

func TestParseSpec_MissingPurpose(t *testing.T) {
	content := `## Requirements
### Requirement: Feature A
The system SHALL do feature A.

#### Scenario: Test
- **WHEN** something
`

	_, err := ParseSpec("test-spec", content)
	if err == nil {
		t.Fatal("expected error for missing Purpose")
	}
	if !strings.Contains(err.Error(), "Purpose") {
		t.Errorf("expected error about Purpose, got %q", err.Error())
	}
}

func TestParseSpec_MissingRequirements(t *testing.T) {
	content := `## Purpose
This is a valid purpose for the specification.
`

	_, err := ParseSpec("test-spec", content)
	if err == nil {
		t.Fatal("expected error for missing Requirements")
	}
	if !strings.Contains(err.Error(), "Requirements") {
		t.Errorf("expected error about Requirements, got %q", err.Error())
	}
}

func TestParseSpec_MultiLineScenarios(t *testing.T) {
	content := `## Purpose
A valid purpose for testing multi-line scenarios.

## Requirements

### Requirement: Multi-line
The system SHALL handle multi-line scenarios.

#### Scenario: Complex flow
- **GIVEN** a precondition
- **WHEN** user performs action
- **THEN** system responds correctly
- **AND** state is updated
`

	spec, err := ParseSpec("test-spec", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(spec.Requirements))
	}
	if len(spec.Requirements[0].Scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(spec.Requirements[0].Scenarios))
	}
	if !strings.Contains(spec.Requirements[0].Scenarios[0].RawText, "GIVEN") {
		t.Error("expected scenario to contain GIVEN")
	}
}

func TestParseChange_Valid(t *testing.T) {
	content := `## Why
This is a detailed explanation of why this change is needed with enough characters.

## What Changes
- **my-spec:** Adds new authentication requirements
`

	change, err := ParseChange("test-change", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if change.Name != "test-change" {
		t.Errorf("expected name 'test-change', got %q", change.Name)
	}
	if len(change.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(change.Deltas))
	}
}

func TestParseChange_MissingWhy(t *testing.T) {
	content := `## What Changes
- **my-spec:** Adds something
`
	_, err := ParseChange("test-change", content)
	if err == nil {
		t.Fatal("expected error for missing Why")
	}
	if !strings.Contains(err.Error(), "Why") {
		t.Errorf("expected error about Why, got %q", err.Error())
	}
}

func TestParseChange_MissingWhatChanges(t *testing.T) {
	content := `## Why
This is a reason for the change.
`
	_, err := ParseChange("test-change", content)
	if err == nil {
		t.Fatal("expected error for missing What Changes")
	}
	if !strings.Contains(err.Error(), "What Changes") {
		t.Errorf("expected error about What Changes, got %q", err.Error())
	}
}

func TestParseChange_NoDeltas(t *testing.T) {
	content := `## Why
This is a reason for the change.

## What Changes
Just some text without any bullet-formatted deltas.
`
	change, err := ParseChange("test-change", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(change.Deltas) != 0 {
		t.Errorf("expected 0 deltas, got %d", len(change.Deltas))
	}
}

func TestParseChange_CRLFLineEndings(t *testing.T) {
	content := "## Why\r\nThis is a detailed reason for the change with enough characters to pass.\r\n\r\n## What Changes\r\n- **my-spec:** Adds new features\r\n"

	change, err := ParseChange("test-change", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(change.Deltas) != 1 {
		t.Errorf("expected 1 delta, got %d", len(change.Deltas))
	}
}
