package commandgen

import (
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/core/globalconfig"
)

func TestGenerateForTool_UnknownTool(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	_, err := GenerateForTool("nonexistent", []string{"propose"}, globalconfig.DeliveryBoth, "1.0.0")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "no adapter registered") {
		t.Errorf("error = %q, expected to contain 'no adapter registered'", err.Error())
	}
}

// skillOnlyAdapter returns skills but no commands.
type skillOnlyAdapter struct {
	mockAdapter
}

func (a *skillOnlyAdapter) GenerateSkills(workflows []string, version string) []CommandContent {
	var results []CommandContent
	for _, wf := range workflows {
		results = append(results, CommandContent{FileName: wf + "-skill.md", Dir: "skills"})
	}
	return results
}

func (a *skillOnlyAdapter) GenerateCommands(workflows []string) []CommandContent {
	var results []CommandContent
	for _, wf := range workflows {
		results = append(results, CommandContent{FileName: wf + "-cmd.md", Dir: "commands"})
	}
	return results
}

func TestGenerateForTool_DeliverySkills(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	Register(&skillOnlyAdapter{mockAdapter: mockAdapter{toolID: "test"}})

	results, err := GenerateForTool("test", []string{"propose"}, globalconfig.DeliverySkills, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Dir != "skills" {
		t.Errorf("expected skills dir, got %q", results[0].Dir)
	}
}

func TestGenerateForTool_DeliveryCommands(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	Register(&skillOnlyAdapter{mockAdapter: mockAdapter{toolID: "test"}})

	results, err := GenerateForTool("test", []string{"propose"}, globalconfig.DeliveryCommands, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Dir != "commands" {
		t.Errorf("expected commands dir, got %q", results[0].Dir)
	}
}

func TestGenerateForTool_DeliveryBoth(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	Register(&skillOnlyAdapter{mockAdapter: mockAdapter{toolID: "test"}})

	results, err := GenerateForTool("test", []string{"propose"}, globalconfig.DeliveryBoth, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (skill+command), got %d", len(results))
	}
}

func TestGenerateForTool_EmptyWorkflows(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	Register(&skillOnlyAdapter{mockAdapter: mockAdapter{toolID: "test"}})

	results, err := GenerateForTool("test", []string{}, globalconfig.DeliveryBoth, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty workflows, got %d", len(results))
	}
}
