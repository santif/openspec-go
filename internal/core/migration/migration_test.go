package migration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santif/openspec-go/internal/core/globalconfig"
)

func TestInferDelivery(t *testing.T) {
	tests := []struct {
		hasSkills   bool
		hasCommands bool
		want        globalconfig.Delivery
	}{
		{true, true, globalconfig.DeliveryBoth},
		{false, true, globalconfig.DeliveryCommands},
		{true, false, globalconfig.DeliverySkills},
		{false, false, globalconfig.DeliverySkills},
	}

	for _, tt := range tests {
		got := InferDelivery(tt.hasSkills, tt.hasCommands)
		if got != tt.want {
			t.Errorf("InferDelivery(%v, %v) = %q, want %q", tt.hasSkills, tt.hasCommands, got, tt.want)
		}
	}
}

func TestScanInstalledWorkflows_Empty(t *testing.T) {
	dir := t.TempDir()
	workflows := ScanInstalledWorkflows(dir)
	if len(workflows) != 0 {
		t.Errorf("expected empty workflows, got %v", workflows)
	}
}

func TestScanInstalledWorkflows_FindsSkills(t *testing.T) {
	dir := t.TempDir()

	// Create a skill file for the "explore" workflow under the claude tool
	skillDir := filepath.Join(dir, ".claude", "skills", "openspec-explore")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644); err != nil {
		t.Fatal(err)
	}

	workflows := ScanInstalledWorkflows(dir)
	if len(workflows) == 0 {
		t.Fatal("expected at least one workflow")
	}

	found := false
	for _, wf := range workflows {
		if wf == "explore" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'explore' in workflows, got %v", workflows)
	}
}

func TestMigrateIfNeeded_NoConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMigrateIfNeeded_ProfileAlreadySet(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	cfg := map[string]interface{}{
		"profile": "core",
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have modified the config
	readData, _ := os.ReadFile(filepath.Join(configDir, "config.json"))
	var result map[string]interface{}
	json.Unmarshal(readData, &result)
	if result["profile"] != "core" {
		t.Error("expected profile to remain 'core'")
	}
}

func TestMigrateIfNeeded_MigratesWhenWorkflowsInstalled(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	// Write config without profile field
	cfg := map[string]interface{}{"featureFlags": map[string]interface{}{}}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a skill file
	skillDir := filepath.Join(dir, ".claude", "skills", "openspec-explore")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644); err != nil {
		t.Fatal(err)
	}

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back config — should now have profile: custom
	readData, _ := os.ReadFile(filepath.Join(configDir, "config.json"))
	var result map[string]interface{}
	json.Unmarshal(readData, &result)
	if result["profile"] != "custom" {
		t.Errorf("expected profile 'custom', got %v", result["profile"])
	}
}

func TestWorkflowToSkillDir(t *testing.T) {
	// Verify key mappings exist
	expected := map[string]string{
		"explore": "openspec-explore",
		"propose": "openspec-propose",
		"archive": "openspec-archive-change",
		"apply":   "openspec-apply-change",
	}

	for wf, dir := range expected {
		got, ok := WorkflowToSkillDir[wf]
		if !ok {
			t.Errorf("missing mapping for workflow %q", wf)
			continue
		}
		if got != dir {
			t.Errorf("WorkflowToSkillDir[%q] = %q, want %q", wf, got, dir)
		}
	}
}
