package artifactgraph

import (
	"strings"
	"testing"
)

func TestParseSchema_Valid(t *testing.T) {
	yaml := `
name: test-schema
version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"
  - id: spec
    generates: "specs/*/spec.md"
    template: "templates/spec.md"
    requires:
      - proposal
`
	schema, err := ParseSchema(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Name != "test-schema" {
		t.Errorf("expected name 'test-schema', got %q", schema.Name)
	}
	if len(schema.Artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(schema.Artifacts))
	}
}

func TestParseSchema_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		errContains string
	}{
		{
			"missing name",
			`version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"`,
			"name is required",
		},
		{
			"missing artifact id",
			`name: test
version: 1
artifacts:
  - generates: "proposal.md"
    template: "templates/proposal.md"`,
			"id is required",
		},
		{
			"missing generates",
			`name: test
version: 1
artifacts:
  - id: proposal
    template: "templates/proposal.md"`,
			"generates is required",
		},
		{
			"missing template",
			`name: test
version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"`,
			"template is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseSchema(tc.yaml)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
			}
		})
	}
}

func TestParseSchema_InvalidVersion(t *testing.T) {
	yaml := `
name: test
version: 0
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for version 0")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("expected error about version, got %q", err.Error())
	}
}

func TestParseSchema_EmptyArtifacts(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts: []
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for empty artifacts")
	}
	if !strings.Contains(err.Error(), "at least one artifact") {
		t.Errorf("expected error about empty artifacts, got %q", err.Error())
	}
}

func TestParseSchema_DuplicateIDs(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"
  - id: proposal
    generates: "other.md"
    template: "templates/other.md"
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for duplicate IDs")
	}
	if !strings.Contains(err.Error(), "Duplicate") {
		t.Errorf("expected error about duplicate, got %q", err.Error())
	}
}

func TestParseSchema_InvalidReference(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"
    requires:
      - nonexistent
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for invalid reference")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected error about invalid reference, got %q", err.Error())
	}
}

func TestParseSchema_SelfCycle(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts:
  - id: proposal
    generates: "proposal.md"
    template: "templates/proposal.md"
    requires:
      - proposal
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for self-cycle")
	}
	if !strings.Contains(err.Error(), "Cyclic") {
		t.Errorf("expected error about cycle, got %q", err.Error())
	}
}

func TestParseSchema_SimpleCycle(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts:
  - id: a
    generates: "a.md"
    template: "templates/a.md"
    requires:
      - b
  - id: b
    generates: "b.md"
    template: "templates/b.md"
    requires:
      - a
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for simple cycle")
	}
	if !strings.Contains(err.Error(), "Cyclic") {
		t.Errorf("expected error about cycle, got %q", err.Error())
	}
}

func TestParseSchema_LongCycle(t *testing.T) {
	yaml := `
name: test
version: 1
artifacts:
  - id: a
    generates: "a.md"
    template: "templates/a.md"
    requires:
      - c
  - id: b
    generates: "b.md"
    template: "templates/b.md"
    requires:
      - a
  - id: c
    generates: "c.md"
    template: "templates/c.md"
    requires:
      - b
`
	_, err := ParseSchema(yaml)
	if err == nil {
		t.Fatal("expected error for long cycle")
	}
	if !strings.Contains(err.Error(), "Cyclic") {
		t.Errorf("expected error about cycle, got %q", err.Error())
	}
}
