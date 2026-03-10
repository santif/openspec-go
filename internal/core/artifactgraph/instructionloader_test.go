package artifactgraph

import (
	"strings"
	"testing"
)

func TestLoadEnrichedInstruction_Basic(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}, Instruction: "Write a proposal document"},
		},
	}
	graph := NewGraphFromSchema(schema)

	result, err := LoadEnrichedInstruction(graph, "proposal", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ArtifactID != "proposal" {
		t.Errorf("expected ArtifactID %q, got %q", "proposal", result.ArtifactID)
	}
	if !strings.Contains(result.Instruction, "Write a proposal document") {
		t.Errorf("expected instruction to contain base text, got %q", result.Instruction)
	}
}

func TestLoadEnrichedInstruction_WithContext(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}, Instruction: "Write a proposal"},
		},
	}
	graph := NewGraphFromSchema(schema)

	result, err := LoadEnrichedInstruction(graph, "proposal", "This is a Go project", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Instruction, "<project-context>") {
		t.Error("expected instruction to contain <project-context> wrapper")
	}
	if !strings.Contains(result.Instruction, "This is a Go project") {
		t.Error("expected instruction to contain project context text")
	}
	if !strings.Contains(result.Instruction, "</project-context>") {
		t.Error("expected instruction to contain </project-context> closing tag")
	}
}

func TestLoadEnrichedInstruction_WithRules(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}, Instruction: "Write a proposal"},
		},
	}
	graph := NewGraphFromSchema(schema)

	rules := map[string][]string{
		"proposal": {"Rule 1", "Rule 2"},
	}
	result, err := LoadEnrichedInstruction(graph, "proposal", "", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Instruction, "<rules>") {
		t.Error("expected instruction to contain <rules> wrapper")
	}
	if !strings.Contains(result.Instruction, "- Rule 1") {
		t.Error("expected instruction to contain Rule 1 as bullet point")
	}
	if !strings.Contains(result.Instruction, "- Rule 2") {
		t.Error("expected instruction to contain Rule 2 as bullet point")
	}
	if !strings.Contains(result.Instruction, "</rules>") {
		t.Error("expected instruction to contain </rules> closing tag")
	}
}

func TestLoadEnrichedInstruction_NotFound(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
	}
	graph := NewGraphFromSchema(schema)

	_, err := LoadEnrichedInstruction(graph, "nonexistent", "", nil)
	if err == nil {
		t.Fatal("expected error for non-existent artifact, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got %q", err.Error())
	}
}

func TestLoadEnrichedInstruction_AllParts(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}, Instruction: "Base instruction text"},
		},
	}
	graph := NewGraphFromSchema(schema)

	rules := map[string][]string{
		"proposal": {"Follow coding standards", "Include tests"},
	}
	result, err := LoadEnrichedInstruction(graph, "proposal", "A microservices project", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Instruction, "Base instruction text") {
		t.Error("expected instruction to contain base instruction text")
	}
	if !strings.Contains(result.Instruction, "<project-context>") {
		t.Error("expected instruction to contain <project-context> wrapper")
	}
	if !strings.Contains(result.Instruction, "A microservices project") {
		t.Error("expected instruction to contain project context")
	}
	if !strings.Contains(result.Instruction, "<rules>") {
		t.Error("expected instruction to contain <rules> wrapper")
	}
	if !strings.Contains(result.Instruction, "- Follow coding standards") {
		t.Error("expected instruction to contain first rule")
	}
	if !strings.Contains(result.Instruction, "- Include tests") {
		t.Error("expected instruction to contain second rule")
	}
}

func TestLoadApplyInstruction_Basic(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
		Apply: &ApplyPhase{Instruction: "Apply these tasks"},
	}
	graph := NewGraphFromSchema(schema)

	result, err := LoadApplyInstruction(graph, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ArtifactID != "apply" {
		t.Errorf("expected ArtifactID %q, got %q", "apply", result.ArtifactID)
	}
	if !strings.Contains(result.Instruction, "Apply these tasks") {
		t.Errorf("expected instruction to contain apply text, got %q", result.Instruction)
	}
}

func TestLoadApplyInstruction_NoApplyPhase(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
	}
	graph := NewGraphFromSchema(schema)

	_, err := LoadApplyInstruction(graph, "", nil)
	if err == nil {
		t.Fatal("expected error when schema has no apply phase, got nil")
	}
	if !strings.Contains(err.Error(), "does not define an apply phase") {
		t.Errorf("expected error to contain 'does not define an apply phase', got %q", err.Error())
	}
}

func TestLoadApplyInstruction_WithContextAndRules(t *testing.T) {
	schema := &SchemaYaml{
		Name:    "test",
		Version: 1,
		Artifacts: []Artifact{
			{ID: "proposal", Generates: "proposal.md", Template: "t.md", Requires: []string{}},
		},
		Apply: &ApplyPhase{Instruction: "Apply the changes"},
	}
	graph := NewGraphFromSchema(schema)

	rules := map[string][]string{
		"apply": {"Run tests after applying", "Update documentation"},
	}
	result, err := LoadApplyInstruction(graph, "Enterprise application context", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Instruction, "<project-context>") {
		t.Error("expected instruction to contain <project-context> wrapper")
	}
	if !strings.Contains(result.Instruction, "Enterprise application context") {
		t.Error("expected instruction to contain project context text")
	}
	if !strings.Contains(result.Instruction, "<rules>") {
		t.Error("expected instruction to contain <rules> wrapper")
	}
	if !strings.Contains(result.Instruction, "- Run tests after applying") {
		t.Error("expected instruction to contain first rule")
	}
	if !strings.Contains(result.Instruction, "- Update documentation") {
		t.Error("expected instruction to contain second rule")
	}
}
