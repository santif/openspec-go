package cli

import (
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/core/validation"
)

// --- titleCase ---

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single lowercase char", "a", "A"},
		{"single uppercase char", "A", "A"},
		{"lowercase word", "changes", "Changes"},
		{"already capitalized", "Changes", "Changes"},
		{"all uppercase", "SPECS", "SPECS"},
		{"multi-word", "hello world", "Hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titleCase(tt.input)
			if got != tt.expected {
				t.Errorf("titleCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- mergeReports ---

func TestMergeReports_BothEmpty(t *testing.T) {
	a := validation.Report{Valid: true}
	b := validation.Report{Valid: true}

	merged := mergeReports(a, b)

	if !merged.Valid {
		t.Error("expected Valid=true when both reports are empty")
	}
	if len(merged.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(merged.Issues))
	}
	if merged.Summary.Errors != 0 || merged.Summary.Warnings != 0 || merged.Summary.Info != 0 {
		t.Errorf("expected all summary counts to be 0, got errors=%d warnings=%d info=%d",
			merged.Summary.Errors, merged.Summary.Warnings, merged.Summary.Info)
	}
}

func TestMergeReports_OneWithErrors(t *testing.T) {
	a := validation.Report{
		Valid: false,
		Issues: []validation.Issue{
			{Level: validation.LevelError, Path: "test", Message: "error1"},
		},
	}
	b := validation.Report{Valid: true}

	merged := mergeReports(a, b)

	if merged.Valid {
		t.Error("expected Valid=false when one report has errors")
	}
	if len(merged.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(merged.Issues))
	}
	if merged.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", merged.Summary.Errors)
	}
}

func TestMergeReports_BothWithMixedIssues(t *testing.T) {
	a := validation.Report{
		Valid: false,
		Issues: []validation.Issue{
			{Level: validation.LevelError, Path: "a", Message: "err"},
			{Level: validation.LevelWarning, Path: "a", Message: "warn"},
		},
	}
	b := validation.Report{
		Valid: true,
		Issues: []validation.Issue{
			{Level: validation.LevelInfo, Path: "b", Message: "info"},
			{Level: validation.LevelWarning, Path: "b", Message: "warn2"},
		},
	}

	merged := mergeReports(a, b)

	if merged.Valid {
		t.Error("expected Valid=false when merged has errors")
	}
	if len(merged.Issues) != 4 {
		t.Errorf("expected 4 issues, got %d", len(merged.Issues))
	}
	if merged.Summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", merged.Summary.Errors)
	}
	if merged.Summary.Warnings != 2 {
		t.Errorf("expected 2 warnings, got %d", merged.Summary.Warnings)
	}
	if merged.Summary.Info != 1 {
		t.Errorf("expected 1 info, got %d", merged.Summary.Info)
	}
}

func TestMergeReports_WarningsOnlyIsValid(t *testing.T) {
	a := validation.Report{
		Valid: true,
		Issues: []validation.Issue{
			{Level: validation.LevelWarning, Path: "a", Message: "warn"},
		},
	}
	b := validation.Report{Valid: true}

	merged := mergeReports(a, b)

	if !merged.Valid {
		t.Error("expected Valid=true when only warnings present")
	}
}

// --- progressBar ---

func TestProgressBar(t *testing.T) {
	// Disable color for deterministic output
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	tests := []struct {
		name      string
		completed int
		total     int
		width     int
		expected  string
	}{
		{"total zero returns dots", 0, 0, 10, ".........."},
		{"completed zero", 0, 5, 10, ".........."},
		{"completed equals total", 5, 5, 10, "##########"},
		{"partial fill", 1, 2, 10, "#####....."},
		{"completed exceeds width clamped", 100, 10, 5, "#####"},
		{"one of four", 1, 4, 8, "##......"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := progressBar(tt.completed, tt.total, tt.width)
			if got != tt.expected {
				t.Errorf("progressBar(%d, %d, %d) = %q, want %q",
					tt.completed, tt.total, tt.width, got, tt.expected)
			}
		})
	}
}

// --- resolveTools ---

func TestResolveTools_Empty(t *testing.T) {
	tools := resolveTools("")
	if tools != nil {
		t.Errorf("expected nil for empty flag, got %d tools", len(tools))
	}
}

func TestResolveTools_None(t *testing.T) {
	tools := resolveTools("none")
	if tools != nil {
		t.Errorf("expected nil for 'none', got %d tools", len(tools))
	}
}

func TestResolveTools_All(t *testing.T) {
	tools := resolveTools("all")
	if len(tools) == 0 {
		t.Fatal("expected at least one tool for 'all'")
	}

	// All returned tools should have a SkillsDir
	for _, tool := range tools {
		if tool.SkillsDir == "" {
			t.Errorf("tool %q returned by 'all' has empty SkillsDir", tool.Value)
		}
	}
}

func TestResolveTools_Specific(t *testing.T) {
	tools := resolveTools("claude,cursor")
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools for 'claude,cursor', got %d", len(tools))
	}

	values := map[string]bool{}
	for _, tool := range tools {
		values[tool.Value] = true
	}
	if !values["claude"] {
		t.Error("expected 'claude' in resolved tools")
	}
	if !values["cursor"] {
		t.Error("expected 'cursor' in resolved tools")
	}
}

func TestResolveTools_SpecificWithSpaces(t *testing.T) {
	tools := resolveTools("claude, cursor")
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools for 'claude, cursor', got %d", len(tools))
	}
}

func TestResolveTools_Nonexistent(t *testing.T) {
	tools := resolveTools("nonexistent")
	if len(tools) != 0 {
		t.Errorf("expected 0 tools for nonexistent value, got %d", len(tools))
	}
}

func TestResolveTools_AllContainsKnownTools(t *testing.T) {
	tools := resolveTools("all")
	values := map[string]bool{}
	for _, tool := range tools {
		values[tool.Value] = true
	}

	// Agents has Available=false, so it should NOT be in "all"
	knownAvailable := []string{"claude", "cursor", "codex"}
	for _, id := range knownAvailable {
		if !values[id] {
			t.Errorf("expected 'all' to include %q", id)
		}
	}

	// The "agents" tool has Available=false and no SkillsDir
	for _, tool := range config.AITools {
		if !tool.Available && values[tool.Value] {
			t.Errorf("tool %q has Available=false but was included in 'all'", tool.Value)
		}
	}
}

// --- replaceConditionalKeywords ---

func TestReplaceConditionalKeywords(t *testing.T) {
	content := "- **WHEN** condition\n- **THEN** outcome\n- **AND** extra"
	cond := &projectconfig.ConditionalsConfig{When: "CUANDO", Then: "ENTONCES", And: "Y"}
	result := replaceConditionalKeywords(content, cond)
	expected := "- **CUANDO** condition\n- **ENTONCES** outcome\n- **Y** extra"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestReplaceConditionalKeywords_NoMatch(t *testing.T) {
	content := "No bold keywords here, just WHEN and THEN"
	cond := &projectconfig.ConditionalsConfig{When: "CUANDO", Then: "ENTONCES", And: "Y"}
	result := replaceConditionalKeywords(content, cond)
	if result != content {
		t.Errorf("expected no changes for non-bold keywords, got %q", result)
	}
}

func TestResolveTools_AllHaveSkillsDir(t *testing.T) {
	tools := resolveTools("all")
	for _, tool := range tools {
		if tool.SkillsDir == "" {
			t.Errorf("tool %q from 'all' has empty SkillsDir", tool.Value)
		}
		if !strings.HasPrefix(tool.SkillsDir, ".") {
			t.Errorf("tool %q SkillsDir %q does not start with '.'", tool.Value, tool.SkillsDir)
		}
	}
}
