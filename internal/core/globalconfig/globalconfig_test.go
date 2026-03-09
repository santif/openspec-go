package globalconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetGlobalConfig_DefaultsWhenFileMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := GetGlobalConfig()

	if cfg.Profile != ProfileCore {
		t.Errorf("expected profile %q, got %q", ProfileCore, cfg.Profile)
	}
	if cfg.Delivery != DeliveryBoth {
		t.Errorf("expected delivery %q, got %q", DeliveryBoth, cfg.Delivery)
	}
	if cfg.FeatureFlags == nil {
		t.Error("expected FeatureFlags to be non-nil empty map")
	}
	if len(cfg.FeatureFlags) != 0 {
		t.Errorf("expected empty FeatureFlags, got %v", cfg.FeatureFlags)
	}
	if cfg.Workflows != nil {
		t.Errorf("expected nil Workflows, got %v", cfg.Workflows)
	}
}

func TestGetGlobalConfig_MergesPartialConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	partial := `{"profile":"custom"}`
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte(partial), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetGlobalConfig()

	if cfg.Profile != ProfileCustom {
		t.Errorf("expected profile %q, got %q", ProfileCustom, cfg.Profile)
	}
	// Other fields should remain at defaults
	if cfg.Delivery != DeliveryBoth {
		t.Errorf("expected delivery default %q, got %q", DeliveryBoth, cfg.Delivery)
	}
	if cfg.FeatureFlags == nil || len(cfg.FeatureFlags) != 0 {
		t.Errorf("expected empty FeatureFlags map, got %v", cfg.FeatureFlags)
	}
}

func TestGetGlobalConfig_HandlesInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	malformed := `{not valid json!!!`
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte(malformed), 0644); err != nil {
		t.Fatal(err)
	}

	// Should return defaults without panicking
	cfg := GetGlobalConfig()

	if cfg.Profile != ProfileCore {
		t.Errorf("expected default profile %q, got %q", ProfileCore, cfg.Profile)
	}
	if cfg.Delivery != DeliveryBoth {
		t.Errorf("expected default delivery %q, got %q", DeliveryBoth, cfg.Delivery)
	}
}

func TestSaveGlobalConfig_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// The openspec config directory should not exist yet
	configDir := filepath.Join(dir, GlobalConfigDirName)
	if _, err := os.Stat(configDir); err == nil {
		t.Fatal("config directory should not exist before save")
	}

	cfg := GlobalConfig{
		Profile:      ProfileCustom,
		Delivery:     DeliverySkills,
		FeatureFlags: map[string]bool{"beta": true},
	}

	if err := SaveGlobalConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Directory should now exist
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("config directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected config path to be a directory")
	}

	// File should exist
	configPath := filepath.Join(configDir, GlobalConfigFileName)
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}

func TestSaveGlobalConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	original := GlobalConfig{
		Profile:      ProfileCustom,
		Delivery:     DeliveryCommands,
		FeatureFlags: map[string]bool{"experimentalFeature": true, "legacy": false},
		Workflows:    []string{"propose", "explore", "apply"},
	}

	if err := SaveGlobalConfig(original); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded := GetGlobalConfig()

	if loaded.Profile != original.Profile {
		t.Errorf("profile: expected %q, got %q", original.Profile, loaded.Profile)
	}
	if loaded.Delivery != original.Delivery {
		t.Errorf("delivery: expected %q, got %q", original.Delivery, loaded.Delivery)
	}
	if len(loaded.FeatureFlags) != len(original.FeatureFlags) {
		t.Errorf("feature flags count: expected %d, got %d", len(original.FeatureFlags), len(loaded.FeatureFlags))
	}
	for k, v := range original.FeatureFlags {
		if loaded.FeatureFlags[k] != v {
			t.Errorf("feature flag %q: expected %v, got %v", k, v, loaded.FeatureFlags[k])
		}
	}
	if len(loaded.Workflows) != len(original.Workflows) {
		t.Errorf("workflows count: expected %d, got %d", len(original.Workflows), len(loaded.Workflows))
	}
	for i, w := range original.Workflows {
		if loaded.Workflows[i] != w {
			t.Errorf("workflow[%d]: expected %q, got %q", i, w, loaded.Workflows[i])
		}
	}

	// Also verify the file is valid JSON with indentation
	configPath := filepath.Join(dir, GlobalConfigDirName, GlobalConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
}

func TestGetGlobalConfigDir_RespectsXDG(t *testing.T) {
	customDir := "/tmp/custom-xdg-config"
	t.Setenv("XDG_CONFIG_HOME", customDir)

	result := GetGlobalConfigDir()

	expected := filepath.Join(customDir, GlobalConfigDirName)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	if !strings.HasPrefix(result, customDir) {
		t.Errorf("expected path to start with XDG_CONFIG_HOME %q, got %q", customDir, result)
	}
}

func TestGetGlobalDataDir_RespectsXDG(t *testing.T) {
	customDir := "/tmp/custom-xdg-data"
	t.Setenv("XDG_DATA_HOME", customDir)

	result := GetGlobalDataDir()

	expected := filepath.Join(customDir, GlobalDataDirName)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	if !strings.HasPrefix(result, customDir) {
		t.Errorf("expected path to start with XDG_DATA_HOME %q, got %q", customDir, result)
	}
}
