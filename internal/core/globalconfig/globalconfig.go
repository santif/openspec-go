package globalconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	GlobalConfigDirName  = "openspec"
	GlobalConfigFileName = "config.json"
	GlobalDataDirName    = "openspec"
)

type Profile string

const (
	ProfileCore   Profile = "core"
	ProfileCustom Profile = "custom"
)

type Delivery string

const (
	DeliveryBoth     Delivery = "both"
	DeliverySkills   Delivery = "skills"
	DeliveryCommands Delivery = "commands"
)

type GlobalConfig struct {
	FeatureFlags map[string]bool `json:"featureFlags,omitempty"`
	Profile      Profile         `json:"profile,omitempty"`
	Delivery     Delivery        `json:"delivery,omitempty"`
	Workflows    []string        `json:"workflows,omitempty"`
}

var defaultConfig = GlobalConfig{
	FeatureFlags: map[string]bool{},
	Profile:      ProfileCore,
	Delivery:     DeliveryBoth,
}

func GetGlobalConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, GlobalConfigDirName)
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, GlobalConfigDirName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", GlobalConfigDirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", GlobalConfigDirName)
}

func GetGlobalDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, GlobalDataDirName)
	}
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, GlobalDataDirName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Local", GlobalDataDirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", GlobalDataDirName)
}

func GetGlobalConfigPath() string {
	return filepath.Join(GetGlobalConfigDir(), GlobalConfigFileName)
}

func GetGlobalConfig() GlobalConfig {
	configPath := GetGlobalConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return copyDefault()
	}

	var parsed GlobalConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid JSON in %s, using defaults\n", configPath)
		return copyDefault()
	}

	// Merge with defaults
	merged := copyDefault()
	if parsed.FeatureFlags != nil {
		for k, v := range parsed.FeatureFlags {
			merged.FeatureFlags[k] = v
		}
	}
	if parsed.Profile != "" {
		merged.Profile = parsed.Profile
	}
	if parsed.Delivery != "" {
		merged.Delivery = parsed.Delivery
	}
	if parsed.Workflows != nil {
		merged.Workflows = parsed.Workflows
	}

	return merged
}

func SaveGlobalConfig(config GlobalConfig) error {
	configDir := GetGlobalConfigDir()
	configPath := GetGlobalConfigPath()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	data = append(data, '\n')
	return os.WriteFile(configPath, data, 0644)
}

func copyDefault() GlobalConfig {
	flags := make(map[string]bool)
	for k, v := range defaultConfig.FeatureFlags {
		flags[k] = v
	}
	return GlobalConfig{
		FeatureFlags: flags,
		Profile:      defaultConfig.Profile,
		Delivery:     defaultConfig.Delivery,
	}
}
