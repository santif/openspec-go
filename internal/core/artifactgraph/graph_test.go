package artifactgraph

import (
	"os"
	"path/filepath"
	"testing"
)

func buildTestSchema(artifacts []Artifact) *SchemaYaml {
	return &SchemaYaml{
		Name:      "test",
		Version:   1,
		Artifacts: artifacts,
	}
}

func TestGetBuildOrder_LinearChain(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "c", Generates: "c.md", Template: "t.md", Requires: []string{"b"}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)
	order := g.GetBuildOrder()

	if len(order) != 3 {
		t.Fatalf("expected 3 items, got %d", len(order))
	}

	// a must come before b, b before c
	indexOf := make(map[string]int)
	for i, id := range order {
		indexOf[id] = i
	}
	if indexOf["a"] >= indexOf["b"] {
		t.Errorf("expected a before b, got %v", order)
	}
	if indexOf["b"] >= indexOf["c"] {
		t.Errorf("expected b before c, got %v", order)
	}
}

func TestGetBuildOrder_DiamondDependency(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "d", Generates: "d.md", Template: "t.md", Requires: []string{"b", "c"}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
		{ID: "c", Generates: "c.md", Template: "t.md", Requires: []string{"a"}},
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)
	order := g.GetBuildOrder()

	if len(order) != 4 {
		t.Fatalf("expected 4 items, got %d", len(order))
	}

	indexOf := make(map[string]int)
	for i, id := range order {
		indexOf[id] = i
	}
	if indexOf["a"] >= indexOf["b"] || indexOf["a"] >= indexOf["c"] {
		t.Errorf("expected a before b and c, got %v", order)
	}
	if indexOf["b"] >= indexOf["d"] || indexOf["c"] >= indexOf["d"] {
		t.Errorf("expected b and c before d, got %v", order)
	}
}

func TestGetBuildOrder_IndependentStableOrder(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "z", Generates: "z.md", Template: "t.md", Requires: []string{}},
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "m", Generates: "m.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)
	order := g.GetBuildOrder()

	if len(order) != 3 {
		t.Fatalf("expected 3 items, got %d", len(order))
	}
	// Independent items should be sorted alphabetically
	if order[0] != "a" || order[1] != "m" || order[2] != "z" {
		t.Errorf("expected sorted order [a, m, z], got %v", order)
	}
}

func TestGetNextArtifacts_Roots(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
	})
	g := NewGraphFromSchema(schema)

	completed := make(map[string]bool)
	next := g.GetNextArtifacts(completed)

	if len(next) != 1 || next[0] != "a" {
		t.Errorf("expected [a], got %v", next)
	}
}

func TestGetNextArtifacts_CompletedDeps(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
		{ID: "c", Generates: "c.md", Template: "t.md", Requires: []string{"a"}},
	})
	g := NewGraphFromSchema(schema)

	completed := map[string]bool{"a": true}
	next := g.GetNextArtifacts(completed)

	if len(next) != 2 {
		t.Fatalf("expected 2 next items, got %d", len(next))
	}
}

func TestGetBlocked_UnmetDependencies(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
		{ID: "c", Generates: "c.md", Template: "t.md", Requires: []string{"a", "b"}},
	})
	g := NewGraphFromSchema(schema)

	completed := make(map[string]bool)
	blocked := g.GetBlocked(completed)

	if _, ok := blocked["b"]; !ok {
		t.Error("expected b to be blocked")
	}
	if _, ok := blocked["c"]; !ok {
		t.Error("expected c to be blocked")
	}
	if _, ok := blocked["a"]; ok {
		t.Error("expected a to NOT be blocked (it's a root)")
	}
}

func TestIsComplete(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
	})
	g := NewGraphFromSchema(schema)

	if g.IsComplete(map[string]bool{"a": true}) {
		t.Error("expected not complete with only a")
	}
	if !g.IsComplete(map[string]bool{"a": true, "b": true}) {
		t.Error("expected complete with both a and b")
	}
}

func TestNewGraphFromYamlContent(t *testing.T) {
	yamlStr := `name: test-schema
version: 1
artifacts:
  - id: proposal
    generates: proposal.md
    template: proposal.md
    requires: []`

	graph, err := NewGraphFromYamlContent(yamlStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if graph.GetName() != "test-schema" {
		t.Errorf("expected name %q, got %q", "test-schema", graph.GetName())
	}
}

func TestNewGraphFromYamlContent_Invalid(t *testing.T) {
	_, err := NewGraphFromYamlContent("not: valid: yaml: [")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestNewGraphFromYaml(t *testing.T) {
	yamlStr := `name: file-schema
version: 1
artifacts:
  - id: spec
    generates: spec.md
    template: spec.md
    requires: []`

	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "schema.yaml")
	if err := os.WriteFile(schemaPath, []byte(yamlStr), 0644); err != nil {
		t.Fatalf("failed to write temp schema file: %v", err)
	}

	graph, err := NewGraphFromYaml(schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if graph.GetName() != "file-schema" {
		t.Errorf("expected name %q, got %q", "file-schema", graph.GetName())
	}
}

func TestNewGraphFromYaml_NotFound(t *testing.T) {
	_, err := NewGraphFromYaml("/nonexistent/path/schema.yaml")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestGetArtifact(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
		{ID: "b", Generates: "b.md", Template: "t.md", Requires: []string{"a"}},
	})
	g := NewGraphFromSchema(schema)

	a := g.GetArtifact("a")
	if a == nil {
		t.Fatal("expected artifact 'a' to be found, got nil")
	}
	if a.ID != "a" {
		t.Errorf("expected artifact ID %q, got %q", "a", a.ID)
	}

	missing := g.GetArtifact("nonexistent")
	if missing != nil {
		t.Errorf("expected nil for non-existent artifact, got %+v", missing)
	}
}

func TestGetName(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)

	if g.GetName() != "test" {
		t.Errorf("expected name %q, got %q", "test", g.GetName())
	}
}

func TestGetVersion(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)

	if g.GetVersion() != 1 {
		t.Errorf("expected version %d, got %d", 1, g.GetVersion())
	}
}

func TestGetSchema(t *testing.T) {
	schema := buildTestSchema([]Artifact{
		{ID: "a", Generates: "a.md", Template: "t.md", Requires: []string{}},
	})
	g := NewGraphFromSchema(schema)

	s := g.GetSchema()
	if s == nil {
		t.Fatal("expected GetSchema() to return non-nil schema")
	}
	if s.Name != "test" {
		t.Errorf("expected schema name %q, got %q", "test", s.Name)
	}
	if s.Version != 1 {
		t.Errorf("expected schema version %d, got %d", 1, s.Version)
	}
}
