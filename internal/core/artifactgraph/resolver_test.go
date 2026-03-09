package artifactgraph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minimalCustomSchemaYAML = `name: custom
version: 1
artifacts:
  - id: proposal
    generates: "changes/*/proposal.md"
    template: proposal.md
`

func TestResolveSchema_BuiltIn(t *testing.T) {
	// Use a temp dir as project root so no project-local schema is found.
	projectRoot := t.TempDir()

	schema, err := ResolveSchema("spec-driven", projectRoot)
	if err != nil {
		t.Fatalf("unexpected error resolving built-in schema: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema for built-in 'spec-driven'")
	}
	if schema.Name != "spec-driven" {
		t.Errorf("expected schema name 'spec-driven', got %q", schema.Name)
	}
	if len(schema.Artifacts) == 0 {
		t.Error("expected at least one artifact in built-in schema")
	}
}

func TestResolveSchema_ProjectLocal(t *testing.T) {
	projectRoot := t.TempDir()

	// Create project-local schema directory and file.
	schemaDir := filepath.Join(projectRoot, "openspec", "schemas", "custom")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create schema dir: %v", err)
	}
	schemaPath := filepath.Join(schemaDir, "schema.yaml")
	if err := os.WriteFile(schemaPath, []byte(minimalCustomSchemaYAML), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	schema, err := ResolveSchema("custom", projectRoot)
	if err != nil {
		t.Fatalf("unexpected error resolving project-local schema: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema for project-local 'custom'")
	}
	if schema.Name != "custom" {
		t.Errorf("expected schema name 'custom', got %q", schema.Name)
	}
	if len(schema.Artifacts) != 1 {
		t.Errorf("expected 1 artifact, got %d", len(schema.Artifacts))
	}
	if schema.Artifacts[0].ID != "proposal" {
		t.Errorf("expected artifact ID 'proposal', got %q", schema.Artifacts[0].ID)
	}
}

func TestResolveSchema_UserOverride(t *testing.T) {
	// Set XDG_DATA_HOME to a temp dir so GetGlobalDataDir() returns a controlled path.
	xdgDataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", xdgDataHome)

	// Create user override schema.
	userSchemaDir := filepath.Join(xdgDataHome, "openspec", "schemas", "custom")
	if err := os.MkdirAll(userSchemaDir, 0755); err != nil {
		t.Fatalf("failed to create user schema dir: %v", err)
	}
	schemaPath := filepath.Join(userSchemaDir, "schema.yaml")
	if err := os.WriteFile(schemaPath, []byte(minimalCustomSchemaYAML), 0644); err != nil {
		t.Fatalf("failed to write user schema file: %v", err)
	}

	// Use an empty project root so project-local is not found.
	projectRoot := t.TempDir()

	schema, err := ResolveSchema("custom", projectRoot)
	if err != nil {
		t.Fatalf("unexpected error resolving user-override schema: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema for user-override 'custom'")
	}
	if schema.Name != "custom" {
		t.Errorf("expected schema name 'custom', got %q", schema.Name)
	}
}

func TestResolveSchema_NotFound(t *testing.T) {
	// Point XDG_DATA_HOME to an empty temp dir to avoid picking up real user data.
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	projectRoot := t.TempDir()

	_, err := ResolveSchema("nonexistent-schema", projectRoot)
	if err == nil {
		t.Fatal("expected error for unknown schema name")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error message, got %q", err.Error())
	}
}

func TestListAvailableSchemas_IncludesBuiltIn(t *testing.T) {
	projectRoot := t.TempDir()

	schemas := ListAvailableSchemas(projectRoot)

	found := false
	for _, s := range schemas {
		if s.Name == "spec-driven" {
			found = true
			if !s.IsBuiltIn {
				t.Error("expected 'spec-driven' to have IsBuiltIn=true")
			}
			if s.Path != "" {
				t.Errorf("expected empty Path for built-in schema, got %q", s.Path)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'spec-driven' to appear in available schemas")
	}
}

func TestListAvailableSchemas_IncludesProjectLocal(t *testing.T) {
	projectRoot := t.TempDir()

	// Create project-local schema.
	schemaDir := filepath.Join(projectRoot, "openspec", "schemas", "my-local")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatalf("failed to create schema dir: %v", err)
	}
	schemaPath := filepath.Join(schemaDir, "schema.yaml")
	localYAML := strings.Replace(minimalCustomSchemaYAML, "name: custom", "name: my-local", 1)
	if err := os.WriteFile(schemaPath, []byte(localYAML), 0644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	schemas := ListAvailableSchemas(projectRoot)

	found := false
	for _, s := range schemas {
		if s.Name == "my-local" {
			found = true
			if s.IsBuiltIn {
				t.Error("expected project-local schema to have IsBuiltIn=false")
			}
			if s.Path == "" {
				t.Error("expected non-empty Path for project-local schema")
			}
			break
		}
	}
	if !found {
		t.Error("expected 'my-local' to appear in available schemas")
	}
}

func TestResolveTemplatePath_BuiltIn(t *testing.T) {
	// Point XDG_DATA_HOME to an empty temp dir to avoid picking up real user data.
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	projectRoot := t.TempDir()

	result := ResolveTemplatePath("spec-driven", "proposal.md", projectRoot)
	if !strings.HasPrefix(result, "builtin:") {
		t.Errorf("expected result to start with 'builtin:', got %q", result)
	}
	if !strings.Contains(result, "spec-driven/templates/proposal.md") {
		t.Errorf("expected result to contain 'spec-driven/templates/proposal.md', got %q", result)
	}
}

func TestResolveTemplatePath_ProjectLocal(t *testing.T) {
	// Point XDG_DATA_HOME to an empty temp dir to avoid picking up real user data.
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	projectRoot := t.TempDir()

	// Create a project-local template file.
	templateDir := filepath.Join(projectRoot, "openspec", "schemas", "my-schema", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template dir: %v", err)
	}
	templatePath := filepath.Join(templateDir, "proposal.md")
	if err := os.WriteFile(templatePath, []byte("# My Template\n"), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	result := ResolveTemplatePath("my-schema", "proposal.md", projectRoot)
	if strings.HasPrefix(result, "builtin:") {
		t.Errorf("expected filesystem path, got built-in marker: %q", result)
	}
	if result != templatePath {
		t.Errorf("expected %q, got %q", templatePath, result)
	}
}

func TestReadTemplate_BuiltIn(t *testing.T) {
	content, err := ReadTemplate("spec-driven", "proposal.md")
	if err != nil {
		t.Fatalf("unexpected error reading built-in template: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty template content")
	}
	if !strings.Contains(content, "## Why") {
		t.Error("expected template to contain '## Why' section")
	}
}
