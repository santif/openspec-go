package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/globalconfig"
	"github.com/santif/openspec-go/internal/core/profiles"
	"github.com/santif/openspec-go/internal/utils"
)

// WorkflowToSkillDir maps workflow IDs to their skill directory names.
var WorkflowToSkillDir = map[string]string{
	"explore":      "openspec-explore",
	"new":          "openspec-new-change",
	"continue":     "openspec-continue-change",
	"apply":        "openspec-apply-change",
	"ff":           "openspec-ff-change",
	"sync":         "openspec-sync-specs",
	"archive":      "openspec-archive-change",
	"bulk-archive": "openspec-bulk-archive-change",
	"verify":       "openspec-verify-change",
	"onboard":      "openspec-onboard",
	"propose":      "openspec-propose",
}

// ScanInstalledWorkflows detects which workflows have skill files installed
// across all configured tools.
func ScanInstalledWorkflows(projectPath string) []string {
	installed := make(map[string]bool)

	for _, tool := range config.AITools {
		if tool.SkillsDir == "" {
			continue
		}
		skillsDir := filepath.Join(projectPath, tool.SkillsDir, "skills")

		for _, workflowID := range profiles.AllWorkflows {
			skillDirName, ok := WorkflowToSkillDir[workflowID]
			if !ok {
				continue
			}
			skillFile := filepath.Join(skillsDir, skillDirName, "SKILL.md")
			if utils.FileExists(skillFile) {
				installed[workflowID] = true
			}
		}
	}

	// Return in AllWorkflows order
	var result []string
	for _, wf := range profiles.AllWorkflows {
		if installed[wf] {
			result = append(result, wf)
		}
	}
	return result
}

// InferDelivery determines the delivery mode based on what's installed.
func InferDelivery(hasSkills, hasCommands bool) globalconfig.Delivery {
	if hasSkills && hasCommands {
		return globalconfig.DeliveryBoth
	}
	if hasCommands {
		return globalconfig.DeliveryCommands
	}
	return globalconfig.DeliverySkills
}

// MigrateIfNeeded performs one-time migration if the global config does not yet
// have a profile field. Called by both init and update before profile resolution.
func MigrateIfNeeded(projectPath string) error {
	configPath := globalconfig.GetGlobalConfigPath()

	var rawConfig map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config file, new user — defaults will apply
		}
		return nil // Can't read config, skip migration
	}

	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil //nolint:nilerr // Invalid JSON, skip migration gracefully
	}

	// If profile is already explicitly set, no migration needed
	if _, hasProfile := rawConfig["profile"]; hasProfile {
		return nil
	}

	// Scan for installed workflows
	installedWorkflows := ScanInstalledWorkflows(projectPath)
	if len(installedWorkflows) == 0 {
		return nil // No workflows installed, new user — defaults will apply
	}

	// Detect if skills or commands are present
	hasSkills := false
	for _, tool := range config.AITools {
		if tool.SkillsDir == "" {
			continue
		}
		skillsDir := filepath.Join(projectPath, tool.SkillsDir, "skills")
		for _, wf := range profiles.AllWorkflows {
			dirName, ok := WorkflowToSkillDir[wf]
			if !ok {
				continue
			}
			if utils.FileExists(filepath.Join(skillsDir, dirName, "SKILL.md")) {
				hasSkills = true
				break
			}
		}
		if hasSkills {
			break
		}
	}

	// Migrate: set profile to custom with detected workflows
	cfg := globalconfig.GetGlobalConfig()
	cfg.Profile = globalconfig.ProfileCustom
	cfg.Workflows = installedWorkflows

	if _, hasDelivery := rawConfig["delivery"]; !hasDelivery {
		cfg.Delivery = InferDelivery(hasSkills, false)
	}

	if err := globalconfig.SaveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("failed to save migrated config: %w", err)
	}

	fmt.Printf("Migrated: custom profile with %d workflows\n", len(installedWorkflows))
	return nil
}
