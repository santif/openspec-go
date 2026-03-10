package schemas

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuiltinSchemas_SchemaYamlReadable(t *testing.T) {
	data, err := BuiltinSchemas.ReadFile("spec-driven/schema.yaml")
	if err != nil {
		t.Fatalf("failed to read spec-driven/schema.yaml: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("spec-driven/schema.yaml is empty")
	}

	// Verify it parses as valid YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("spec-driven/schema.yaml is not valid YAML: %v", err)
	}

	// Verify key fields exist
	if _, ok := parsed["name"]; !ok {
		t.Error("schema.yaml missing 'name' field")
	}
	if _, ok := parsed["artifacts"]; !ok {
		t.Error("schema.yaml missing 'artifacts' field")
	}
}

func TestBuiltinSchemas_AllTemplatesExist(t *testing.T) {
	templates := []struct {
		name string
		path string
	}{
		{"proposal", "spec-driven/templates/proposal.md"},
		{"spec", "spec-driven/templates/spec.md"},
		{"design", "spec-driven/templates/design.md"},
		{"tasks", "spec-driven/templates/tasks.md"},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			data, err := BuiltinSchemas.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.path, err)
			}
			if len(data) == 0 {
				t.Errorf("%s is empty", tt.path)
			}
		})
	}
}

func TestBuiltinSchemas_TemplatesHaveExpectedSections(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			"proposal has Why section",
			"spec-driven/templates/proposal.md",
			[]string{"## Why", "## What Changes"},
		},
		{
			"spec has ADDED section",
			"spec-driven/templates/spec.md",
			[]string{"ADDED"},
		},
		{
			"design has Context section",
			"spec-driven/templates/design.md",
			[]string{"Context"},
		},
		{
			"tasks has checkbox format",
			"spec-driven/templates/tasks.md",
			[]string{"- [ ]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := BuiltinSchemas.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.path, err)
			}
			content := string(data)
			for _, substr := range tt.contains {
				if !strings.Contains(content, substr) {
					t.Errorf("%s does not contain expected text %q", tt.path, substr)
				}
			}
		})
	}
}
