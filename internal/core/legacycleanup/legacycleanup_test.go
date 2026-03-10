package legacycleanup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/utils"
)

func TestHasOpenSpecMarkers(t *testing.T) {
	content := config.OpenSpecMarkers.Start + "\nsome content\n" + config.OpenSpecMarkers.End
	if !HasOpenSpecMarkers(content) {
		t.Error("expected markers to be detected")
	}

	if HasOpenSpecMarkers("no markers here") {
		t.Error("expected no markers detected")
	}

	if HasOpenSpecMarkers(config.OpenSpecMarkers.Start + " only start") {
		t.Error("expected false when only start marker present")
	}
}

func TestDetectLegacyArtifacts_ConfigFiles(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md with OpenSpec markers
	claudeContent := "Before\n" + config.OpenSpecMarkers.Start + "\nOpenSpec content\n" + config.OpenSpecMarkers.End + "\nAfter"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(claudeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create CLINE.md without markers
	if err := os.WriteFile(filepath.Join(dir, "CLINE.md"), []byte("No markers"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ConfigFiles) != 1 || result.ConfigFiles[0] != "CLAUDE.md" {
		t.Errorf("expected [CLAUDE.md], got %v", result.ConfigFiles)
	}
	if !result.HasLegacyArtifacts {
		t.Error("expected HasLegacyArtifacts to be true")
	}
}

func TestDetectLegacyArtifacts_SlashCommandDirs(t *testing.T) {
	dir := t.TempDir()

	// Create a legacy slash command directory
	cmdDir := filepath.Join(dir, ".claude", "commands", "openspec")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, d := range result.SlashCommandDirs {
		if d == ".claude/commands/openspec" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected .claude/commands/openspec in dirs, got %v", result.SlashCommandDirs)
	}
}

func TestDetectLegacyArtifacts_SlashCommandFiles(t *testing.T) {
	dir := t.TempDir()

	// Create file-based legacy commands
	cursorDir := filepath.Join(dir, ".cursor", "commands")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cursorDir, "openspec-explore.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range result.SlashCommandFiles {
		if f == ".cursor/commands/openspec-explore.md" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected .cursor/commands/openspec-explore.md in files, got %v", result.SlashCommandFiles)
	}
}

func TestDetectLegacyArtifacts_StructureFiles(t *testing.T) {
	dir := t.TempDir()
	openspecDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(openspecDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create openspec/AGENTS.md
	if err := os.WriteFile(filepath.Join(openspecDir, "AGENTS.md"), []byte("agents"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create openspec/project.md
	if err := os.WriteFile(filepath.Join(openspecDir, "project.md"), []byte("project"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasOpenspecAgents {
		t.Error("expected HasOpenspecAgents to be true")
	}
	if !result.HasProjectMd {
		t.Error("expected HasProjectMd to be true")
	}
}

func TestDetectLegacyArtifacts_NoLegacy(t *testing.T) {
	dir := t.TempDir()

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HasLegacyArtifacts {
		t.Error("expected no legacy artifacts")
	}
}

func TestCleanupLegacyArtifacts_RemovesMarkers(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md with markers
	content := "Before\n" + config.OpenSpecMarkers.Start + "\nOpenSpec\n" + config.OpenSpecMarkers.End + "\nAfter"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	detection := &LegacyDetectionResult{
		ConfigFilesToUpdate: []string{"CLAUDE.md"},
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ModifiedFiles) != 1 {
		t.Errorf("expected 1 modified file, got %d", len(result.ModifiedFiles))
	}

	// Verify markers were removed but file still exists
	newContent, err := utils.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if strings.Contains(newContent, config.OpenSpecMarkers.Start) {
		t.Error("expected start marker to be removed")
	}
	if !strings.Contains(newContent, "Before") {
		t.Error("expected 'Before' content preserved")
	}
	if !strings.Contains(newContent, "After") {
		t.Error("expected 'After' content preserved")
	}
}

func TestCleanupLegacyArtifacts_DeletesDirs(t *testing.T) {
	dir := t.TempDir()

	cmdDir := filepath.Join(dir, ".claude", "commands", "openspec")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "cmd.md"), []byte("cmd"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := &LegacyDetectionResult{
		SlashCommandDirs: []string{".claude/commands/openspec"},
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DeletedDirs) != 1 {
		t.Errorf("expected 1 deleted dir, got %d", len(result.DeletedDirs))
	}

	if utils.DirectoryExists(cmdDir) {
		t.Error("expected directory to be deleted")
	}
}

func TestCleanupLegacyArtifacts_DeletesFiles(t *testing.T) {
	dir := t.TempDir()

	cursorDir := filepath.Join(dir, ".cursor", "commands")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(cursorDir, "openspec-explore.md")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := &LegacyDetectionResult{
		SlashCommandFiles: []string{".cursor/commands/openspec-explore.md"},
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DeletedFiles) != 1 {
		t.Errorf("expected 1 deleted file, got %d", len(result.DeletedFiles))
	}

	if utils.FileExists(file) {
		t.Error("expected file to be deleted")
	}
}

func TestCleanupLegacyArtifacts_DeletesOpenspecAgents(t *testing.T) {
	dir := t.TempDir()
	openspecDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(openspecDir, 0755); err != nil {
		t.Fatal(err)
	}
	agentsPath := filepath.Join(openspecDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("agents"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := &LegacyDetectionResult{
		HasOpenspecAgents: true,
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range result.DeletedFiles {
		if f == "openspec/AGENTS.md" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected openspec/AGENTS.md in deleted files, got %v", result.DeletedFiles)
	}

	if utils.FileExists(agentsPath) {
		t.Error("expected openspec/AGENTS.md to be deleted")
	}
}

func TestFormatDetectionSummary_Empty(t *testing.T) {
	detection := &LegacyDetectionResult{}
	summary := FormatDetectionSummary(detection)
	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}

func TestFormatDetectionSummary_WithArtifacts(t *testing.T) {
	detection := &LegacyDetectionResult{
		ConfigFilesToUpdate: []string{"CLAUDE.md"},
		SlashCommandDirs:    []string{".claude/commands/openspec"},
		HasProjectMd:        true,
	}

	summary := FormatDetectionSummary(detection)
	if !strings.Contains(summary, "Upgrading") {
		t.Error("expected upgrade header")
	}
	if !strings.Contains(summary, "CLAUDE.md") {
		t.Error("expected CLAUDE.md mentioned")
	}
	if !strings.Contains(summary, ".claude/commands/openspec") {
		t.Error("expected slash command dir mentioned")
	}
	if !strings.Contains(summary, "project.md") {
		t.Error("expected project.md migration hint")
	}
}

func TestFormatCleanupSummary(t *testing.T) {
	result := &CleanupResult{
		DeletedFiles:  []string{"openspec/AGENTS.md"},
		DeletedDirs:   []string{".claude/commands/openspec"},
		ModifiedFiles: []string{"CLAUDE.md"},
	}

	summary := FormatCleanupSummary(result)
	if !strings.Contains(summary, "Cleaned up") {
		t.Error("expected cleanup header")
	}
	if !strings.Contains(summary, "AGENTS.md") {
		t.Error("expected deleted file mentioned")
	}
	if !strings.Contains(summary, "CLAUDE.md") {
		t.Error("expected modified file mentioned")
	}
}

func TestGetToolsFromLegacyArtifacts(t *testing.T) {
	detection := &LegacyDetectionResult{
		SlashCommandDirs:  []string{".claude/commands/openspec"},
		SlashCommandFiles: []string{".cursor/commands/openspec-explore.md"},
	}

	tools := GetToolsFromLegacyArtifacts(detection)
	toolMap := make(map[string]bool)
	for _, t := range tools {
		toolMap[t] = true
	}

	if !toolMap["claude"] {
		t.Error("expected 'claude' in tools")
	}
	if !toolMap["cursor"] {
		t.Error("expected 'cursor' in tools")
	}
}

func TestGetToolsFromLegacyArtifacts_Empty(t *testing.T) {
	detection := &LegacyDetectionResult{}
	tools := GetToolsFromLegacyArtifacts(detection)
	if len(tools) != 0 {
		t.Errorf("expected no tools, got %v", tools)
	}
}

func TestCleanupLegacyArtifacts_ErrorReadingConfigFile(t *testing.T) {
	dir := t.TempDir()

	// Point to a config file that doesn't exist anymore
	detection := &LegacyDetectionResult{
		ConfigFilesToUpdate: []string{"nonexistent.md"},
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
	}
	if !strings.Contains(result.Errors[0], "Failed to read") {
		t.Errorf("expected read error, got: %s", result.Errors[0])
	}
}

func TestCleanupLegacyArtifacts_ErrorDeletingFile(t *testing.T) {
	dir := t.TempDir()

	// Point to a file that doesn't exist
	detection := &LegacyDetectionResult{
		SlashCommandFiles: []string{"nonexistent/file.md"},
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestCleanupLegacyArtifacts_ProjectMdMigration(t *testing.T) {
	dir := t.TempDir()

	detection := &LegacyDetectionResult{
		HasProjectMd: true,
	}

	result, err := CleanupLegacyArtifacts(dir, detection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.ProjectMdNeedsMigration {
		t.Error("expected ProjectMdNeedsMigration to be true")
	}
}

func TestFormatCleanupSummary_WithErrors(t *testing.T) {
	result := &CleanupResult{
		Errors: []string{"Failed to delete something"},
	}

	summary := FormatCleanupSummary(result)
	if !strings.Contains(summary, "Errors during cleanup") {
		t.Error("expected errors section in summary")
	}
	if !strings.Contains(summary, "Warning:") {
		t.Error("expected Warning prefix for errors")
	}
}

func TestFormatCleanupSummary_WithProjectMdMigration(t *testing.T) {
	result := &CleanupResult{
		DeletedFiles:            []string{"file.md"},
		ProjectMdNeedsMigration: true,
	}

	summary := FormatCleanupSummary(result)
	if !strings.Contains(summary, "project.md") {
		t.Error("expected project.md migration hint")
	}
	if !strings.Contains(summary, "config.yaml") {
		t.Error("expected config.yaml mention in migration hint")
	}
}

func TestFormatCleanupSummary_Empty(t *testing.T) {
	result := &CleanupResult{}
	summary := FormatCleanupSummary(result)
	if summary != "" {
		t.Errorf("expected empty summary, got %q", summary)
	}
}

func TestDetectLegacyArtifacts_RootAgentsWithMarkers(t *testing.T) {
	dir := t.TempDir()

	// Create root AGENTS.md with markers
	agentsContent := "Before\n" + config.OpenSpecMarkers.Start + "\nOpenSpec content\n" + config.OpenSpecMarkers.End + "\nAfter"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.HasRootAgentsWithMarkers {
		t.Error("expected HasRootAgentsWithMarkers to be true")
	}
	if !result.HasLegacyArtifacts {
		t.Error("expected HasLegacyArtifacts to be true")
	}
	// AGENTS.md is also in LegacyConfigFiles, so it should show up as a ConfigFile too
	foundConfig := false
	for _, f := range result.ConfigFiles {
		if f == "AGENTS.md" {
			foundConfig = true
			break
		}
	}
	if !foundConfig {
		t.Error("expected AGENTS.md in ConfigFiles")
	}
}

func TestDetectLegacyArtifacts_MultipleSlashCommandFiles(t *testing.T) {
	dir := t.TempDir()

	// Create multiple file-based legacy commands for different tools
	cursorDir := filepath.Join(dir, ".cursor", "commands")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cursorDir, "openspec-explore.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cursorDir, "openspec-propose.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectLegacyArtifacts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.SlashCommandFiles) < 2 {
		t.Errorf("expected at least 2 slash command files, got %d: %v", len(result.SlashCommandFiles), result.SlashCommandFiles)
	}
}

func TestFormatDetectionSummary_OnlyConfigFiles(t *testing.T) {
	detection := &LegacyDetectionResult{
		ConfigFilesToUpdate: []string{"CLAUDE.md", "CLINE.md"},
	}

	summary := FormatDetectionSummary(detection)
	if !strings.Contains(summary, "Files to update") {
		t.Error("expected 'Files to update' section")
	}
	if !strings.Contains(summary, "CLAUDE.md") {
		t.Error("expected CLAUDE.md in summary")
	}
	if !strings.Contains(summary, "CLINE.md") {
		t.Error("expected CLINE.md in summary")
	}
}

func TestFormatDetectionSummary_OnlyProjectMd(t *testing.T) {
	detection := &LegacyDetectionResult{
		HasProjectMd: true,
	}

	summary := FormatDetectionSummary(detection)
	if !strings.Contains(summary, "project.md") {
		t.Error("expected project.md migration hint")
	}
}
