package artifactgraph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectCompleted_MissingDir(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
	}
	g := NewGraphFromSchema(schema)

	completed := DetectCompleted(g, "/nonexistent/path")
	if len(completed) != 0 {
		t.Errorf("expected 0 completed, got %d", len(completed))
	}
}

func TestDetectCompleted_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
	}
	g := NewGraphFromSchema(schema)

	completed := DetectCompleted(g, dir)
	if len(completed) != 0 {
		t.Errorf("expected 0 completed, got %d", len(completed))
	}
}

func TestDetectCompleted_FileExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
	}
	g := NewGraphFromSchema(schema)

	completed := DetectCompleted(g, dir)
	if !completed["proposal"] {
		t.Error("expected proposal to be completed")
	}
}

func TestDetectCompleted_GlobPattern(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", "auth")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "spec", Generates: "specs/*/spec.md", Template: "t.md", Requires: []string{}},
		},
	}
	g := NewGraphFromSchema(schema)

	completed := DetectCompleted(g, dir)
	if !completed["spec"] {
		t.Error("expected spec to be completed via glob")
	}
}

func TestDetectCompleted_MixedCompletion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
			{ID: "design", Generates: "design.md", Template: "t.md", Requires: []string{"proposal"}},
		},
	}
	g := NewGraphFromSchema(schema)

	completed := DetectCompleted(g, dir)
	if !completed["proposal"] {
		t.Error("expected proposal to be completed")
	}
	if completed["design"] {
		t.Error("expected design to NOT be completed")
	}
}
