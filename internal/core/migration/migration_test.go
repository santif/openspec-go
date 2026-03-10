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

func TestMigrateIfNeeded_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte("{broken json!!!"), 0644); err != nil {
		t.Fatal(err)
	}

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("expected nil error for invalid JSON (graceful skip), got: %v", err)
	}
}

func TestMigrateIfNeeded_NoWorkflowsInstalled(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	// Config without profile (triggers migration path) but no workflows installed
	cfg := map[string]interface{}{"featureFlags": map[string]interface{}{}}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Config should NOT have been modified (no workflows found)
	readData, _ := os.ReadFile(filepath.Join(configDir, "config.json"))
	var result map[string]interface{}
	json.Unmarshal(readData, &result)
	if _, hasProfile := result["profile"]; hasProfile {
		t.Error("expected profile to NOT be set when no workflows are installed")
	}
}

func TestMigrateIfNeeded_SetsDeliveryWhenMissing(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))

	// Config without profile or delivery fields
	cfg := map[string]interface{}{"featureFlags": map[string]interface{}{}}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a skill file so migration triggers
	skillDir := filepath.Join(dir, ".claude", "skills", "openspec-propose")
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

	// Read back and verify delivery was inferred
	readData, _ := os.ReadFile(filepath.Join(configDir, "config.json"))
	var result map[string]interface{}
	json.Unmarshal(readData, &result)
	if result["delivery"] == nil {
		t.Error("expected delivery to be set after migration")
	}
	if result["profile"] != "custom" {
		t.Errorf("expected profile 'custom', got %v", result["profile"])
	}
}

func TestScanInstalledWorkflows_MultipleTools(t *testing.T) {
	dir := t.TempDir()

	// Create skills for multiple tools
	for _, toolDir := range []string{".claude", ".cursor"} {
		skillDir := filepath.Join(dir, toolDir, "skills", "openspec-explore")
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Also add a different workflow
	applyDir := filepath.Join(dir, ".claude", "skills", "openspec-apply-change")
	if err := os.MkdirAll(applyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(applyDir, "SKILL.md"), []byte("skill"), 0644); err != nil {
		t.Fatal(err)
	}

	workflows := ScanInstalledWorkflows(dir)
	if len(workflows) < 2 {
		t.Fatalf("expected at least 2 workflows, got %d: %v", len(workflows), workflows)
	}

	foundExplore, foundApply := false, false
	for _, wf := range workflows {
		if wf == "explore" {
			foundExplore = true
		}
		if wf == "apply" {
			foundApply = true
		}
	}
	if !foundExplore {
		t.Error("expected 'explore' in workflows")
	}
	if !foundApply {
		t.Error("expected 'apply' in workflows")
	}
}

func TestMigrateIfNeeded_ConfigReadError(t *testing.T) {
	dir := t.TempDir()
	// Point to a non-existent XDG path (no config file at all)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "nonexistent"))

	err := MigrateIfNeeded(dir)
	if err != nil {
		t.Fatalf("expected nil error when config cannot be read, got: %v", err)
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
