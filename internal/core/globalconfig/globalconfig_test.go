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

func TestGetGlobalConfigDir_FallbackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	result := GetGlobalConfigDir()

	if !strings.Contains(result, filepath.Join(".config", GlobalConfigDirName)) {
		t.Errorf("expected path containing '.config/%s', got %q", GlobalConfigDirName, result)
	}
}

func TestGetGlobalDataDir_FallbackToHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	result := GetGlobalDataDir()

	if !strings.Contains(result, filepath.Join(".local", "share", GlobalDataDirName)) {
		t.Errorf("expected path containing '.local/share/%s', got %q", GlobalDataDirName, result)
	}
}

func TestGetGlobalConfig_MergesFeatureFlags(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := GlobalConfig{
		Profile:      ProfileCore,
		Delivery:     DeliveryBoth,
		FeatureFlags: map[string]bool{"experimental": true, "legacy": false},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded := GetGlobalConfig()

	if len(loaded.FeatureFlags) != 2 {
		t.Fatalf("expected 2 feature flags, got %d: %v", len(loaded.FeatureFlags), loaded.FeatureFlags)
	}
	if !loaded.FeatureFlags["experimental"] {
		t.Error("expected 'experimental' feature flag to be true")
	}
	if loaded.FeatureFlags["legacy"] {
		t.Error("expected 'legacy' feature flag to be false")
	}
}

func TestGetGlobalConfig_MergesWorkflows(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	partial := `{"workflows":["propose","explore"]}`
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte(partial), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetGlobalConfig()

	if len(cfg.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(cfg.Workflows))
	}
	if cfg.Workflows[0] != "propose" || cfg.Workflows[1] != "explore" {
		t.Errorf("unexpected workflows: %v", cfg.Workflows)
	}
	// Defaults should still be applied for other fields
	if cfg.Profile != ProfileCore {
		t.Errorf("expected default profile %q, got %q", ProfileCore, cfg.Profile)
	}
}

func TestGetGlobalConfig_MergesDelivery(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	partial := `{"delivery":"skills"}`
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte(partial), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetGlobalConfig()

	if cfg.Delivery != DeliverySkills {
		t.Errorf("expected delivery %q, got %q", DeliverySkills, cfg.Delivery)
	}
	if cfg.Profile != ProfileCore {
		t.Errorf("expected default profile, got %q", cfg.Profile)
	}
}

func TestSaveGlobalConfig_FileEndsWithNewline(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := GlobalConfig{
		Profile:      ProfileCore,
		Delivery:     DeliveryBoth,
		FeatureFlags: map[string]bool{},
	}

	if err := SaveGlobalConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configPath := filepath.Join(dir, GlobalConfigDirName, GlobalConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("config file is empty")
	}
	if data[len(data)-1] != '\n' {
		t.Error("expected config file to end with newline")
	}
}

func TestSaveGlobalConfig_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := GlobalConfig{
		Profile:      ProfileCustom,
		Delivery:     DeliveryCommands,
		FeatureFlags: map[string]bool{"a": true},
		Workflows:    []string{"explore"},
	}

	if err := SaveGlobalConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configPath := filepath.Join(dir, GlobalConfigDirName, GlobalConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved config is not valid JSON: %v", err)
	}

	// Verify indentation (should be 2 spaces)
	if !strings.Contains(string(data), "  \"") {
		t.Error("expected JSON to be indented with 2 spaces")
	}
}

func TestGetGlobalConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path := GetGlobalConfigPath()
	expected := filepath.Join(dir, GlobalConfigDirName, GlobalConfigFileName)
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestCopyDefault_ReturnsIndependentCopy(t *testing.T) {
	a := copyDefault()
	b := copyDefault()

	a.FeatureFlags["test"] = true
	if b.FeatureFlags["test"] {
		t.Error("modifying one copy should not affect the other")
	}

	a.Profile = ProfileCustom
	if a.Profile != ProfileCustom {
		t.Error("expected profile to be set to ProfileCustom")
	}
	if b.Profile != ProfileCore {
		t.Error("modifying profile on one copy should not affect the other")
	}
}

func TestGetGlobalConfigDir_NoXDG_NoWindows(t *testing.T) {
	// Ensure XDG is empty so fallback path is used
	t.Setenv("XDG_CONFIG_HOME", "")

	result := GetGlobalConfigDir()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", GlobalConfigDirName)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetGlobalDataDir_NoXDG_NoWindows(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	result := GetGlobalDataDir()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "share", GlobalDataDirName)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSaveGlobalConfig_ErrorOnReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	// Create a file where the config directory should be, so MkdirAll fails
	blocker := filepath.Join(dir, GlobalConfigDirName)
	if err := os.WriteFile(blocker, []byte("not a dir"), 0444); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := GlobalConfig{Profile: ProfileCore}
	err := SaveGlobalConfig(cfg)
	if err == nil {
		t.Error("expected error when config dir cannot be created")
	}
}

func TestGetGlobalConfig_OverridesDelivery(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Config with delivery override
	partial := `{"delivery":"commands","featureFlags":{"beta":true}}`
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte(partial), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetGlobalConfig()
	if cfg.Delivery != DeliveryCommands {
		t.Errorf("expected delivery %q, got %q", DeliveryCommands, cfg.Delivery)
	}
	if !cfg.FeatureFlags["beta"] {
		t.Error("expected beta feature flag to be true")
	}
}

func TestGetGlobalConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, GlobalConfigDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Empty JSON object should merge with defaults
	if err := os.WriteFile(filepath.Join(configDir, GlobalConfigFileName), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := GetGlobalConfig()

	if cfg.Profile != ProfileCore {
		t.Errorf("expected default profile %q, got %q", ProfileCore, cfg.Profile)
	}
	if cfg.Delivery != DeliveryBoth {
		t.Errorf("expected default delivery %q, got %q", DeliveryBoth, cfg.Delivery)
	}
}
