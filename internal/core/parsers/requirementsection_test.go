package parsers

import (
	"testing"
)

func TestExtractRequirementsSection_Exists(t *testing.T) {
	content := `# My Spec

## Purpose
Some purpose here.

## Requirements

### Requirement: Auth
The system SHALL authenticate users.

#### Scenario: Login
- **WHEN** user enters credentials
- **THEN** system validates them

### Requirement: Logout
The system SHALL allow users to logout.

#### Scenario: Session end
- **WHEN** user clicks logout
- **THEN** session is destroyed
`

	parts := ExtractRequirementsSection(content)

	if parts.HeaderLine != "## Requirements" {
		t.Errorf("expected HeaderLine '## Requirements', got %q", parts.HeaderLine)
	}
	if len(parts.BodyBlocks) != 2 {
		t.Fatalf("expected 2 body blocks, got %d", len(parts.BodyBlocks))
	}
	if parts.BodyBlocks[0].Name != "Auth" {
		t.Errorf("expected first block name 'Auth', got %q", parts.BodyBlocks[0].Name)
	}
	if parts.BodyBlocks[1].Name != "Logout" {
		t.Errorf("expected second block name 'Logout', got %q", parts.BodyBlocks[1].Name)
	}
}

func TestExtractRequirementsSection_NotExists(t *testing.T) {
	content := `# My Spec

## Purpose
Some purpose here.
`

	parts := ExtractRequirementsSection(content)

	if parts.HeaderLine != "## Requirements" {
		t.Errorf("expected HeaderLine '## Requirements', got %q", parts.HeaderLine)
	}
	if len(parts.BodyBlocks) != 0 {
		t.Errorf("expected 0 body blocks, got %d", len(parts.BodyBlocks))
	}
	if parts.Before == "" {
		t.Error("expected Before to contain original content")
	}
}

func TestExtractRequirementsSection_PreservesOrder(t *testing.T) {
	content := `## Requirements

### Requirement: Zebra
SHALL do Z.

#### Scenario: Z
- **WHEN** Z

### Requirement: Alpha
SHALL do A.

#### Scenario: A
- **WHEN** A

### Requirement: Middle
SHALL do M.

#### Scenario: M
- **WHEN** M
`

	parts := ExtractRequirementsSection(content)

	if len(parts.BodyBlocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(parts.BodyBlocks))
	}
	expected := []string{"Zebra", "Alpha", "Middle"}
	for i, name := range expected {
		if parts.BodyBlocks[i].Name != name {
			t.Errorf("block %d: expected %q, got %q", i, name, parts.BodyBlocks[i].Name)
		}
	}
}

func TestExtractRequirementsSection_WithPreamble(t *testing.T) {
	content := `## Requirements

This is a preamble describing the requirements.

### Requirement: First
SHALL do something.

#### Scenario: Test
- **WHEN** test
`

	parts := ExtractRequirementsSection(content)

	if parts.Preamble == "" {
		t.Error("expected non-empty preamble")
	}
	if len(parts.BodyBlocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(parts.BodyBlocks))
	}
}

func TestExtractRequirementsSection_CaseInsensitive(t *testing.T) {
	content := `## requirements

### Requirement: Test
SHALL test.

#### Scenario: T
- **WHEN** T
`

	parts := ExtractRequirementsSection(content)

	if parts.HeaderLine != "## requirements" {
		t.Errorf("expected original casing preserved in HeaderLine, got %q", parts.HeaderLine)
	}
	if len(parts.BodyBlocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(parts.BodyBlocks))
	}
}

func TestExtractRequirementsSection_WithAfterContent(t *testing.T) {
	content := `## Requirements

### Requirement: First
SHALL do something.

#### Scenario: Test
- **WHEN** test

## Appendix

Some extra content here.
`

	parts := ExtractRequirementsSection(content)

	if len(parts.BodyBlocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(parts.BodyBlocks))
	}
	if parts.After == "" {
		t.Error("expected non-empty After section")
	}
}
