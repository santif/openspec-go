package parsers

import (
	"testing"
)

func TestParseDeltaSpec_AllSections(t *testing.T) {
	content := `## ADDED Requirements

### Requirement: New Feature
The system SHALL add a new feature.

#### Scenario: Basic
- **WHEN** something happens
- **THEN** feature works

## MODIFIED Requirements

### Requirement: Existing Feature
The system SHALL modify an existing feature.

#### Scenario: Updated
- **WHEN** updated
- **THEN** works differently

## REMOVED Requirements

### Requirement: Old Feature

## RENAMED Requirements

- FROM: ` + "`### Requirement: Legacy Name`" + `
- TO: ` + "`### Requirement: Modern Name`" + `
`

	plan := ParseDeltaSpec(content)

	if len(plan.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(plan.Added))
	}
	if len(plan.Modified) != 1 {
		t.Errorf("expected 1 modified, got %d", len(plan.Modified))
	}
	if len(plan.Removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(plan.Removed))
	}
	if len(plan.Renamed) != 1 {
		t.Errorf("expected 1 renamed, got %d", len(plan.Renamed))
	}
	if plan.Renamed[0].From != "Legacy Name" {
		t.Errorf("expected renamed from 'Legacy Name', got %q", plan.Renamed[0].From)
	}
	if plan.Renamed[0].To != "Modern Name" {
		t.Errorf("expected renamed to 'Modern Name', got %q", plan.Renamed[0].To)
	}
}

func TestParseDeltaSpec_CaseInsensitive(t *testing.T) {
	content := `## added requirements

### Requirement: Feature
The system SHALL do something.

#### Scenario: Test
- **WHEN** test
`

	plan := ParseDeltaSpec(content)

	if len(plan.Added) != 1 {
		t.Errorf("expected 1 added, got %d", len(plan.Added))
	}
	if !plan.SectionPresence.Added {
		t.Error("expected SectionPresence.Added to be true")
	}
}

func TestParseDeltaSpec_SectionPresence(t *testing.T) {
	content := `## ADDED Requirements

### Requirement: Feature
SHALL do something.

#### Scenario: T
- **WHEN** test
`

	plan := ParseDeltaSpec(content)

	if !plan.SectionPresence.Added {
		t.Error("expected Added section presence")
	}
	if plan.SectionPresence.Modified {
		t.Error("expected no Modified section presence")
	}
	if plan.SectionPresence.Removed {
		t.Error("expected no Removed section presence")
	}
	if plan.SectionPresence.Renamed {
		t.Error("expected no Renamed section presence")
	}
}

func TestParseRemovedNames_Headers(t *testing.T) {
	content := `## REMOVED Requirements

### Requirement: Feature A
### Requirement: Feature B
`

	plan := ParseDeltaSpec(content)

	if len(plan.Removed) != 2 {
		t.Fatalf("expected 2 removed, got %d", len(plan.Removed))
	}
	if plan.Removed[0] != "Feature A" {
		t.Errorf("expected 'Feature A', got %q", plan.Removed[0])
	}
	if plan.Removed[1] != "Feature B" {
		t.Errorf("expected 'Feature B', got %q", plan.Removed[1])
	}
}

func TestParseRemovedNames_BulletList(t *testing.T) {
	content := `## REMOVED Requirements

- ` + "`### Requirement: Feature A`" + `
- ` + "`### Requirement: Feature B`" + `
`

	plan := ParseDeltaSpec(content)

	if len(plan.Removed) != 2 {
		t.Fatalf("expected 2 removed, got %d", len(plan.Removed))
	}
	if plan.Removed[0] != "Feature A" {
		t.Errorf("expected 'Feature A', got %q", plan.Removed[0])
	}
}

func TestParseRenamedPairs(t *testing.T) {
	content := `## RENAMED Requirements

- FROM: ` + "`### Requirement: Old Name`" + `
- TO: ` + "`### Requirement: New Name`" + `

- FROM: ` + "`### Requirement: Another Old`" + `
- TO: ` + "`### Requirement: Another New`" + `
`

	plan := ParseDeltaSpec(content)

	if len(plan.Renamed) != 2 {
		t.Fatalf("expected 2 renamed pairs, got %d", len(plan.Renamed))
	}
	if plan.Renamed[0].From != "Old Name" {
		t.Errorf("expected from 'Old Name', got %q", plan.Renamed[0].From)
	}
	if plan.Renamed[0].To != "New Name" {
		t.Errorf("expected to 'New Name', got %q", plan.Renamed[0].To)
	}
}

func TestNormalizeRequirementName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Feature A  ", "Feature A"},
		{"Feature A", "Feature A"},
		{"\tSpaced\t", "Spaced"},
	}

	for _, tc := range tests {
		got := NormalizeRequirementName(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeRequirementName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
