package profiledrift

import (
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/globalconfig"
	"github.com/santif/openspec-go/internal/core/profiles"
	"github.com/santif/openspec-go/internal/utils"
)

// toKnownWorkflows filters a list of workflow strings to only those in AllWorkflows.
func toKnownWorkflows(workflows []string) []string {
	known := make(map[string]bool)
	for _, wf := range profiles.AllWorkflows {
		known[wf] = true
	}

	var result []string
	for _, wf := range workflows {
		if known[wf] {
			result = append(result, wf)
		}
	}
	return result
}

// HasToolProfileOrDeliveryDrift detects if a single tool has profile/delivery drift
// against the desired state. It checks:
// - required skill artifacts missing for selected workflows
// - skill artifacts that exist for deselected workflows
func HasToolProfileOrDeliveryDrift(
	projectPath string,
	toolID string,
	desiredWorkflows []string,
	delivery globalconfig.Delivery,
) bool {
	var tool *config.AIToolOption
	for i := range config.AITools {
		if config.AITools[i].Value == toolID {
			tool = &config.AITools[i]
			break
		}
	}
	if tool == nil || tool.SkillsDir == "" {
		return false
	}

	knownDesired := toKnownWorkflows(desiredWorkflows)
	desiredSet := make(map[string]bool)
	for _, wf := range knownDesired {
		desiredSet[wf] = true
	}

	skillsDir := filepath.Join(projectPath, tool.SkillsDir, "skills")
	shouldGenerateSkills := delivery != globalconfig.DeliveryCommands

	if shouldGenerateSkills {
		// Check that all desired workflows have skill files
		for _, wf := range knownDesired {
			dirName, ok := config.WorkflowToSkillDir[wf]
			if !ok {
				continue
			}
			skillFile := filepath.Join(skillsDir, dirName, "SKILL.md")
			if !utils.FileExists(skillFile) {
				return true
			}
		}

		// Check for deselected workflows that still have artifacts
		for _, wf := range profiles.AllWorkflows {
			if desiredSet[wf] {
				continue
			}
			dirName, ok := config.WorkflowToSkillDir[wf]
			if !ok {
				continue
			}
			skillDir := filepath.Join(skillsDir, dirName)
			if utils.DirectoryExists(skillDir) {
				return true
			}
		}
	} else {
		// Skills delivery is off — any skill artifacts are drift
		for _, wf := range profiles.AllWorkflows {
			dirName, ok := config.WorkflowToSkillDir[wf]
			if !ok {
				continue
			}
			skillDir := filepath.Join(skillsDir, dirName)
			if utils.DirectoryExists(skillDir) {
				return true
			}
		}
	}

	return false
}

// GetToolsNeedingProfileSync returns configured tools that currently need a profile/delivery sync.
func GetToolsNeedingProfileSync(
	projectPath string,
	desiredWorkflows []string,
	delivery globalconfig.Delivery,
	configuredTools []string,
) []string {
	var needSync []string
	for _, toolID := range configuredTools {
		if HasToolProfileOrDeliveryDrift(projectPath, toolID, desiredWorkflows, delivery) {
			needSync = append(needSync, toolID)
		}
	}
	return needSync
}

// GetConfiguredTools returns tools that have a skills directory on disk.
func GetConfiguredTools(projectPath string) []string {
	var tools []string
	for _, tool := range config.AITools {
		if tool.SkillsDir == "" {
			continue
		}
		toolDir := filepath.Join(projectPath, tool.SkillsDir)
		if utils.DirectoryExists(toolDir) {
			tools = append(tools, tool.Value)
		}
	}
	return tools
}
