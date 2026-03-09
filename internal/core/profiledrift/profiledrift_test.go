package profiledrift

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santif/openspec-go/internal/core/globalconfig"
)

func createSkillFile(t *testing.T, projectPath, toolDir, skillDirName string) {
	t.Helper()
	dir := filepath.Join(projectPath, toolDir, "skills", skillDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("skill"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestHasToolProfileOrDeliveryDrift_NoDrift(t *testing.T) {
	dir := t.TempDir()

	// Install exactly the desired workflows
	createSkillFile(t, dir, ".claude", "openspec-explore")
	createSkillFile(t, dir, ".claude", "openspec-propose")

	drift := HasToolProfileOrDeliveryDrift(dir, "claude", []string{"explore", "propose"}, globalconfig.DeliverySkills)
	if drift {
		t.Error("expected no drift when all desired skills are installed")
	}
}

func TestHasToolProfileOrDeliveryDrift_MissingSkill(t *testing.T) {
	dir := t.TempDir()

	// Install only one of two desired workflows
	createSkillFile(t, dir, ".claude", "openspec-explore")

	drift := HasToolProfileOrDeliveryDrift(dir, "claude", []string{"explore", "propose"}, globalconfig.DeliverySkills)
	if !drift {
		t.Error("expected drift when a desired skill is missing")
	}
}

func TestHasToolProfileOrDeliveryDrift_ExtraSkill(t *testing.T) {
	dir := t.TempDir()

	// Install desired + one extra
	createSkillFile(t, dir, ".claude", "openspec-explore")
	createSkillFile(t, dir, ".claude", "openspec-archive-change")

	drift := HasToolProfileOrDeliveryDrift(dir, "claude", []string{"explore"}, globalconfig.DeliverySkills)
	if !drift {
		t.Error("expected drift when deselected workflow has artifacts")
	}
}

func TestHasToolProfileOrDeliveryDrift_CommandsOnlyDelivery(t *testing.T) {
	dir := t.TempDir()

	// Skills exist but delivery is commands-only
	createSkillFile(t, dir, ".claude", "openspec-explore")

	drift := HasToolProfileOrDeliveryDrift(dir, "claude", []string{"explore"}, globalconfig.DeliveryCommands)
	if !drift {
		t.Error("expected drift when delivery is commands-only but skills exist")
	}
}

func TestHasToolProfileOrDeliveryDrift_UnknownTool(t *testing.T) {
	dir := t.TempDir()

	drift := HasToolProfileOrDeliveryDrift(dir, "nonexistent-tool", []string{"explore"}, globalconfig.DeliverySkills)
	if drift {
		t.Error("expected no drift for unknown tool")
	}
}

func TestGetToolsNeedingProfileSync(t *testing.T) {
	dir := t.TempDir()

	// Claude has drift (missing skill), cursor does not exist
	createSkillFile(t, dir, ".claude", "openspec-explore")

	tools := GetToolsNeedingProfileSync(dir, []string{"explore", "propose"}, globalconfig.DeliverySkills, []string{"claude"})
	if len(tools) != 1 || tools[0] != "claude" {
		t.Errorf("expected [claude], got %v", tools)
	}
}

func TestGetToolsNeedingProfileSync_NoDrift(t *testing.T) {
	dir := t.TempDir()

	createSkillFile(t, dir, ".claude", "openspec-explore")

	tools := GetToolsNeedingProfileSync(dir, []string{"explore"}, globalconfig.DeliverySkills, []string{"claude"})
	if len(tools) != 0 {
		t.Errorf("expected no tools needing sync, got %v", tools)
	}
}

func TestGetConfiguredTools(t *testing.T) {
	dir := t.TempDir()

	// Create .claude directory
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}

	tools := GetConfiguredTools(dir)
	found := false
	for _, tool := range tools {
		if tool == "claude" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'claude' in configured tools, got %v", tools)
	}
}

func TestGetConfiguredTools_Empty(t *testing.T) {
	dir := t.TempDir()
	tools := GetConfiguredTools(dir)
	if len(tools) != 0 {
		t.Errorf("expected no configured tools, got %v", tools)
	}
}

func TestToKnownWorkflows(t *testing.T) {
	result := toKnownWorkflows([]string{"explore", "unknown", "propose", "invalid"})
	expected := map[string]bool{"explore": true, "propose": true}

	if len(result) != 2 {
		t.Fatalf("expected 2 workflows, got %d: %v", len(result), result)
	}
	for _, wf := range result {
		if !expected[wf] {
			t.Errorf("unexpected workflow %q", wf)
		}
	}
}
