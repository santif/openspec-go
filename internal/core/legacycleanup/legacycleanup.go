package legacycleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/utils"
)

// LegacyConfigFiles are config files at the project root that may contain OpenSpec markers.
var LegacyConfigFiles = []string{
	"CLAUDE.md",
	"CLINE.md",
	"CODEBUDDY.md",
	"COSTRICT.md",
	"QODER.md",
	"IFLOW.md",
	"AGENTS.md",
	"QWEN.md",
}

// LegacySlashCommandPattern describes how a tool stored legacy slash commands.
type LegacySlashCommandPattern struct {
	Type     string   // "directory" or "files"
	Path     string   // for directory type
	Patterns []string // for files type (glob patterns)
}

// LegacySlashCommandPaths maps tool IDs to their legacy slash command patterns.
var LegacySlashCommandPaths = map[string]LegacySlashCommandPattern{
	"claude":         {Type: "directory", Path: ".claude/commands/openspec"},
	"codebuddy":      {Type: "directory", Path: ".codebuddy/commands/openspec"},
	"qoder":          {Type: "directory", Path: ".qoder/commands/openspec"},
	"crush":          {Type: "directory", Path: ".crush/commands/openspec"},
	"gemini":         {Type: "directory", Path: ".gemini/commands/openspec"},
	"costrict":       {Type: "directory", Path: ".cospec/openspec/commands"},
	"cursor":         {Type: "files", Patterns: []string{".cursor/commands/openspec-*.md"}},
	"windsurf":       {Type: "files", Patterns: []string{".windsurf/workflows/openspec-*.md"}},
	"kilocode":       {Type: "files", Patterns: []string{".kilocode/workflows/openspec-*.md"}},
	"kiro":           {Type: "files", Patterns: []string{".kiro/prompts/openspec-*.prompt.md"}},
	"github-copilot": {Type: "files", Patterns: []string{".github/prompts/openspec-*.prompt.md"}},
	"amazon-q":       {Type: "files", Patterns: []string{".amazonq/prompts/openspec-*.md"}},
	"cline":          {Type: "files", Patterns: []string{".clinerules/workflows/openspec-*.md"}},
	"roocode":        {Type: "files", Patterns: []string{".roo/commands/openspec-*.md"}},
	"auggie":         {Type: "files", Patterns: []string{".augment/commands/openspec-*.md"}},
	"factory":        {Type: "files", Patterns: []string{".factory/commands/openspec-*.md"}},
	"opencode":       {Type: "files", Patterns: []string{".opencode/command/opsx-*.md", ".opencode/command/openspec-*.md"}},
	"continue":       {Type: "files", Patterns: []string{".continue/prompts/openspec-*.prompt"}},
	"antigravity":    {Type: "files", Patterns: []string{".agent/workflows/openspec-*.md"}},
	"iflow":          {Type: "files", Patterns: []string{".iflow/commands/openspec-*.md"}},
	"qwen":           {Type: "files", Patterns: []string{".qwen/commands/openspec-*.toml"}},
	"codex":          {Type: "files", Patterns: []string{".codex/prompts/openspec-*.md"}},
}

// LegacyDetectionResult holds the results of scanning for legacy artifacts.
type LegacyDetectionResult struct {
	ConfigFiles             []string
	ConfigFilesToUpdate     []string
	SlashCommandDirs        []string
	SlashCommandFiles       []string
	HasOpenspecAgents       bool
	HasProjectMd            bool
	HasRootAgentsWithMarkers bool
	HasLegacyArtifacts      bool
}

// CleanupResult holds the results of a cleanup operation.
type CleanupResult struct {
	DeletedFiles            []string
	ModifiedFiles           []string
	DeletedDirs             []string
	ProjectMdNeedsMigration bool
	Errors                  []string
}

// HasOpenSpecMarkers checks if content contains both OpenSpec start and end markers.
func HasOpenSpecMarkers(content string) bool {
	return strings.Contains(content, config.OpenSpecMarkers.Start) &&
		strings.Contains(content, config.OpenSpecMarkers.End)
}

// DetectLegacyArtifacts scans a project for legacy OpenSpec artifacts.
func DetectLegacyArtifacts(projectPath string) (*LegacyDetectionResult, error) {
	result := &LegacyDetectionResult{}

	// Detect legacy config files
	for _, fileName := range LegacyConfigFiles {
		filePath := filepath.Join(projectPath, fileName)
		if !utils.FileExists(filePath) {
			continue
		}
		content, err := utils.ReadFile(filePath)
		if err != nil {
			continue
		}
		if HasOpenSpecMarkers(content) {
			result.ConfigFiles = append(result.ConfigFiles, fileName)
			result.ConfigFilesToUpdate = append(result.ConfigFilesToUpdate, fileName)
		}
	}

	// Detect legacy slash commands
	for _, pattern := range LegacySlashCommandPaths {
		if pattern.Type == "directory" && pattern.Path != "" {
			dirPath := filepath.Join(projectPath, pattern.Path)
			if utils.DirectoryExists(dirPath) {
				result.SlashCommandDirs = append(result.SlashCommandDirs, pattern.Path)
			}
		} else if pattern.Type == "files" {
			for _, p := range pattern.Patterns {
				files := findLegacySlashCommandFiles(projectPath, p)
				result.SlashCommandFiles = append(result.SlashCommandFiles, files...)
			}
		}
	}

	// Detect legacy structure files
	openspecAgentsPath := filepath.Join(projectPath, "openspec", "AGENTS.md")
	result.HasOpenspecAgents = utils.FileExists(openspecAgentsPath)

	projectMdPath := filepath.Join(projectPath, "openspec", "project.md")
	result.HasProjectMd = utils.FileExists(projectMdPath)

	rootAgentsPath := filepath.Join(projectPath, "AGENTS.md")
	if utils.FileExists(rootAgentsPath) {
		content, err := utils.ReadFile(rootAgentsPath)
		if err == nil {
			result.HasRootAgentsWithMarkers = HasOpenSpecMarkers(content)
		}
	}

	result.HasLegacyArtifacts = len(result.ConfigFiles) > 0 ||
		len(result.SlashCommandDirs) > 0 ||
		len(result.SlashCommandFiles) > 0 ||
		result.HasOpenspecAgents ||
		result.HasRootAgentsWithMarkers ||
		result.HasProjectMd

	return result, nil
}

// findLegacySlashCommandFiles finds files matching a glob-like pattern.
func findLegacySlashCommandFiles(projectPath, pattern string) []string {
	var found []string

	lastSlash := strings.LastIndex(pattern, "/")
	if lastSlash == -1 {
		return found
	}
	dirPart := pattern[:lastSlash]
	filePart := pattern[lastSlash+1:]

	dirPath := filepath.Join(projectPath, dirPart)
	if !utils.DirectoryExists(dirPath) {
		return found
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return found
	}

	// Convert glob to regex: escape special chars, replace * with .*
	regexPattern := regexp.QuoteMeta(filePart)
	regexPattern = strings.ReplaceAll(regexPattern, `\*`, `.*`)
	re, err := regexp.Compile("^" + regexPattern + "$")
	if err != nil {
		return found
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if re.MatchString(entry.Name()) {
			found = append(found, dirPart+"/"+entry.Name())
		}
	}

	return found
}

// CleanupLegacyArtifacts removes legacy OpenSpec artifacts from a project.
// Config files are never deleted — only OpenSpec markers are removed from them.
func CleanupLegacyArtifacts(projectPath string, detection *LegacyDetectionResult) (*CleanupResult, error) {
	result := &CleanupResult{
		ProjectMdNeedsMigration: detection.HasProjectMd,
	}

	// Remove marker blocks from config files (NEVER delete config files)
	for _, fileName := range detection.ConfigFilesToUpdate {
		filePath := filepath.Join(projectPath, fileName)
		content, err := utils.ReadFile(filePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to read %s: %v", fileName, err))
			continue
		}
		newContent := utils.RemoveMarkerBlock(content, config.OpenSpecMarkers.Start, config.OpenSpecMarkers.End)
		if err := utils.WriteFile(filePath, newContent); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to modify %s: %v", fileName, err))
			continue
		}
		result.ModifiedFiles = append(result.ModifiedFiles, fileName)
	}

	// Delete legacy slash command directories
	for _, dirPath := range detection.SlashCommandDirs {
		fullPath := filepath.Join(projectPath, dirPath)
		if err := os.RemoveAll(fullPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete directory %s: %v", dirPath, err))
			continue
		}
		result.DeletedDirs = append(result.DeletedDirs, dirPath)
	}

	// Delete legacy slash command files
	for _, filePath := range detection.SlashCommandFiles {
		fullPath := filepath.Join(projectPath, filePath)
		if err := os.Remove(fullPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete %s: %v", filePath, err))
			continue
		}
		result.DeletedFiles = append(result.DeletedFiles, filePath)
	}

	// Delete openspec/AGENTS.md
	if detection.HasOpenspecAgents {
		agentsPath := filepath.Join(projectPath, "openspec", "AGENTS.md")
		if utils.FileExists(agentsPath) {
			if err := os.Remove(agentsPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete openspec/AGENTS.md: %v", err))
			} else {
				result.DeletedFiles = append(result.DeletedFiles, "openspec/AGENTS.md")
			}
		}
	}

	return result, nil
}

// FormatDetectionSummary generates a summary of detected legacy artifacts.
func FormatDetectionSummary(detection *LegacyDetectionResult) string {
	var lines []string

	var removals []string
	for _, dir := range detection.SlashCommandDirs {
		removals = append(removals, dir+"/")
	}
	removals = append(removals, detection.SlashCommandFiles...)
	if detection.HasOpenspecAgents {
		removals = append(removals, "openspec/AGENTS.md")
	}

	if len(removals) == 0 && len(detection.ConfigFilesToUpdate) == 0 && !detection.HasProjectMd {
		return ""
	}

	lines = append(lines, "Upgrading to the new OpenSpec")
	lines = append(lines, "")
	lines = append(lines, "OpenSpec now uses agent skills, the emerging standard across coding")
	lines = append(lines, "agents. This simplifies your setup while keeping everything working")
	lines = append(lines, "as before.")
	lines = append(lines, "")

	if len(removals) > 0 {
		lines = append(lines, "Files to remove")
		lines = append(lines, "No user content to preserve:")
		for _, path := range removals {
			lines = append(lines, fmt.Sprintf("  - %s", path))
		}
	}

	if len(detection.ConfigFilesToUpdate) > 0 {
		if len(removals) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "Files to update")
		lines = append(lines, "OpenSpec markers will be removed, your content preserved:")
		for _, file := range detection.ConfigFilesToUpdate {
			lines = append(lines, fmt.Sprintf("  - %s", file))
		}
	}

	if detection.HasProjectMd {
		if len(removals) > 0 || len(detection.ConfigFilesToUpdate) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, formatProjectMdMigrationHint())
	}

	return strings.Join(lines, "\n")
}

// FormatCleanupSummary generates a summary of cleanup actions taken.
func FormatCleanupSummary(result *CleanupResult) string {
	var lines []string

	if len(result.DeletedFiles) > 0 || len(result.DeletedDirs) > 0 || len(result.ModifiedFiles) > 0 {
		lines = append(lines, "Cleaned up legacy files:")
		for _, file := range result.DeletedFiles {
			lines = append(lines, fmt.Sprintf("  Removed %s", file))
		}
		for _, dir := range result.DeletedDirs {
			lines = append(lines, fmt.Sprintf("  Removed %s/ (replaced by /opsx:*)", dir))
		}
		for _, file := range result.ModifiedFiles {
			lines = append(lines, fmt.Sprintf("  Removed OpenSpec markers from %s", file))
		}
	}

	if result.ProjectMdNeedsMigration {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, formatProjectMdMigrationHint())
	}

	if len(result.Errors) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "Errors during cleanup:")
		for _, e := range result.Errors {
			lines = append(lines, fmt.Sprintf("  Warning: %s", e))
		}
	}

	return strings.Join(lines, "\n")
}

// GetToolsFromLegacyArtifacts extracts tool IDs from detected legacy artifacts.
func GetToolsFromLegacyArtifacts(detection *LegacyDetectionResult) []string {
	tools := make(map[string]bool)

	for _, dir := range detection.SlashCommandDirs {
		for toolID, pattern := range LegacySlashCommandPaths {
			if pattern.Type == "directory" && pattern.Path == dir {
				tools[toolID] = true
				break
			}
		}
	}

	for _, file := range detection.SlashCommandFiles {
		normalizedFile := strings.ReplaceAll(file, "\\", "/")
		for toolID, pattern := range LegacySlashCommandPaths {
			if pattern.Type != "files" {
				continue
			}
			matched := false
			for _, p := range pattern.Patterns {
				regexPattern := regexp.QuoteMeta(p)
				regexPattern = strings.ReplaceAll(regexPattern, `\*`, `.*`)
				re, err := regexp.Compile("^" + regexPattern + "$")
				if err != nil {
					continue
				}
				if re.MatchString(normalizedFile) {
					tools[toolID] = true
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
	}

	var result []string
	for id := range tools {
		result = append(result, id)
	}
	return result
}

func formatProjectMdMigrationHint() string {
	lines := []string{
		"Needs your attention",
		"  - openspec/project.md",
		"    We won't delete this file. It may contain useful project context.",
		"",
		"    The new openspec/config.yaml has a \"context:\" section for planning",
		"    context. This is included in every OpenSpec request and works more",
		"    reliably than the old project.md approach.",
		"",
		"    Review project.md, move any useful content to config.yaml's context",
		"    section, then delete the file when ready.",
	}
	return strings.Join(lines, "\n")
}
