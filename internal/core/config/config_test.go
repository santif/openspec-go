package config

import (
	"strings"
	"testing"
)

// --- OpenSpecDirName ---

func TestOpenSpecDirName(t *testing.T) {
	if OpenSpecDirName != "openspec" {
		t.Errorf("OpenSpecDirName = %q, want %q", OpenSpecDirName, "openspec")
	}
}

// --- OpenSpecMarkers ---

func TestOpenSpecMarkers_AreHTMLComments(t *testing.T) {
	if !strings.HasPrefix(OpenSpecMarkers.Start, "<!--") || !strings.HasSuffix(OpenSpecMarkers.Start, "-->") {
		t.Errorf("Start marker %q is not a valid HTML comment", OpenSpecMarkers.Start)
	}
	if !strings.HasPrefix(OpenSpecMarkers.End, "<!--") || !strings.HasSuffix(OpenSpecMarkers.End, "-->") {
		t.Errorf("End marker %q is not a valid HTML comment", OpenSpecMarkers.End)
	}
}

func TestOpenSpecMarkers_ContainExpectedText(t *testing.T) {
	if !strings.Contains(OpenSpecMarkers.Start, "OPENSPEC:START") {
		t.Errorf("Start marker %q does not contain 'OPENSPEC:START'", OpenSpecMarkers.Start)
	}
	if !strings.Contains(OpenSpecMarkers.End, "OPENSPEC:END") {
		t.Errorf("End marker %q does not contain 'OPENSPEC:END'", OpenSpecMarkers.End)
	}
}

// --- AITools ---

func TestAITools_NoDuplicateValues(t *testing.T) {
	seen := make(map[string]bool)
	for _, tool := range AITools {
		if seen[tool.Value] {
			t.Errorf("duplicate AITool Value: %q", tool.Value)
		}
		seen[tool.Value] = true
	}
}

func TestAITools_AllFieldsPopulated(t *testing.T) {
	for _, tool := range AITools {
		if tool.Name == "" {
			t.Errorf("AITool with Value=%q has empty Name", tool.Value)
		}
		if tool.Value == "" {
			t.Errorf("AITool with Name=%q has empty Value", tool.Name)
		}
		if tool.SuccessLabel == "" {
			t.Errorf("AITool %q has empty SuccessLabel", tool.Value)
		}
	}
}

func TestAITools_SkillsDirsStartWithDot(t *testing.T) {
	for _, tool := range AITools {
		if tool.SkillsDir != "" && !strings.HasPrefix(tool.SkillsDir, ".") {
			t.Errorf("AITool %q SkillsDir %q does not start with '.'", tool.Value, tool.SkillsDir)
		}
	}
}

func TestAITools_KnownToolsExist(t *testing.T) {
	known := map[string]bool{
		"claude": false,
		"cursor": false,
		"codex":  false,
	}

	for _, tool := range AITools {
		if _, ok := known[tool.Value]; ok {
			known[tool.Value] = true
		}
	}

	for id, found := range known {
		if !found {
			t.Errorf("expected known tool %q to exist in AITools", id)
		}
	}
}

func TestAITools_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, tool := range AITools {
		if seen[tool.Name] {
			t.Errorf("duplicate AITool Name: %q", tool.Name)
		}
		seen[tool.Name] = true
	}
}

// --- WorkflowToSkillDir ---

func TestWorkflowToSkillDir_AllValuesNonEmpty(t *testing.T) {
	for key, val := range WorkflowToSkillDir {
		if val == "" {
			t.Errorf("WorkflowToSkillDir[%q] has empty value", key)
		}
	}
}

func TestWorkflowToSkillDir_AllStartWithOpenspec(t *testing.T) {
	for key, val := range WorkflowToSkillDir {
		if !strings.HasPrefix(val, "openspec-") {
			t.Errorf("WorkflowToSkillDir[%q] = %q, does not start with 'openspec-'", key, val)
		}
	}
}

func TestWorkflowToSkillDir_KnownWorkflows(t *testing.T) {
	known := []string{"explore", "propose", "apply", "archive"}
	for _, wf := range known {
		if _, ok := WorkflowToSkillDir[wf]; !ok {
			t.Errorf("expected workflow %q to exist in WorkflowToSkillDir", wf)
		}
	}
}
