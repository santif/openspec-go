package commandgen

import (
	"strings"
	"testing"
)

func TestSkillTemplate_KnownWorkflow(t *testing.T) {
	tmpl := SkillTemplate("propose")
	if tmpl.Name != "openspec-propose" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "openspec-propose")
	}
	if tmpl.Description == "" {
		t.Error("expected non-empty Description")
	}
	if tmpl.Instructions == "" {
		t.Error("expected non-empty Instructions")
	}
	if tmpl.License != "MIT" {
		t.Errorf("License = %q, want %q", tmpl.License, "MIT")
	}
	if tmpl.Author != "openspec" {
		t.Errorf("Author = %q, want %q", tmpl.Author, "openspec")
	}
}

func TestSkillTemplate_UnknownWorkflow(t *testing.T) {
	tmpl := SkillTemplate("nonexistent")
	if tmpl.Name != "openspec-nonexistent" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "openspec-nonexistent")
	}
	if !strings.Contains(tmpl.Description, "nonexistent") {
		t.Errorf("Description = %q, expected to contain 'nonexistent'", tmpl.Description)
	}
	if tmpl.Instructions == "" {
		t.Error("expected non-empty fallback Instructions")
	}
}

func TestCommandTemplate_KnownWorkflow(t *testing.T) {
	tmpl := CommandTemplate("propose")
	if tmpl.ID != "propose" {
		t.Errorf("ID = %q, want %q", tmpl.ID, "propose")
	}
	if tmpl.Name != "OPSX: Propose" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "OPSX: Propose")
	}
	if tmpl.Description == "" {
		t.Error("expected non-empty Description")
	}
	if tmpl.Category != "Workflow" {
		t.Errorf("Category = %q, want %q", tmpl.Category, "Workflow")
	}
	if len(tmpl.Tags) == 0 {
		t.Error("expected non-empty Tags")
	}
	if tmpl.Body == "" {
		t.Error("expected non-empty Body")
	}
}

func TestCommandTemplate_UnknownWorkflow(t *testing.T) {
	tmpl := CommandTemplate("nonexistent")
	if tmpl.ID != "nonexistent" {
		t.Errorf("ID = %q, want %q", tmpl.ID, "nonexistent")
	}
	if tmpl.Name != "OPSX: nonexistent" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "OPSX: nonexistent")
	}
	if !strings.Contains(tmpl.Description, "nonexistent") {
		t.Errorf("Description = %q, expected to contain 'nonexistent'", tmpl.Description)
	}
}

func TestGenerateSkillContent(t *testing.T) {
	tmpl := SkillTemplateData{
		Name:         "test-skill",
		Description:  "A test skill",
		Instructions: "# Test\n\nDo the thing.\n",
	}
	content := GenerateSkillContent(tmpl, "1.2.3")

	checks := []struct {
		name, substr string
	}{
		{"frontmatter start", "---\n"},
		{"name field", "name: test-skill"},
		{"description field", "description: A test skill"},
		{"license default", "license: MIT"},
		{"author default", "author: openspec"},
		{"version default", `version: "1.0"`},
		{"generatedBy", `generatedBy: "1.2.3"`},
		{"instructions body", "# Test"},
	}
	for _, check := range checks {
		if !strings.Contains(content, check.substr) {
			t.Errorf("GenerateSkillContent missing %s: %q not found in output", check.name, check.substr)
		}
	}
}

func TestGenerateSkillContent_WithExplicitFields(t *testing.T) {
	tmpl := SkillTemplateData{
		Name:          "custom",
		Description:   "Custom skill",
		Instructions:  "body",
		License:       "Apache-2.0",
		Compatibility: "Custom compat",
		Author:        "someone",
		Version:       "2.0",
	}
	content := GenerateSkillContent(tmpl, "0.1.0")

	if !strings.Contains(content, "license: Apache-2.0") {
		t.Error("expected custom license")
	}
	if !strings.Contains(content, "author: someone") {
		t.Error("expected custom author")
	}
	if !strings.Contains(content, `version: "2.0"`) {
		t.Error("expected custom version")
	}
}

func TestSkillTemplate_AllKnownWorkflows(t *testing.T) {
	// Verify all known workflows return non-empty templates
	knownWorkflows := []string{"propose", "explore", "apply", "archive", "new", "continue", "ff", "sync", "bulk-archive", "verify", "onboard"}
	for _, wf := range knownWorkflows {
		tmpl := SkillTemplate(wf)
		if tmpl.Name == "" {
			t.Errorf("SkillTemplate(%q): expected non-empty Name", wf)
		}
		if tmpl.Description == "" {
			t.Errorf("SkillTemplate(%q): expected non-empty Description", wf)
		}
		if tmpl.Instructions == "" {
			t.Errorf("SkillTemplate(%q): expected non-empty Instructions", wf)
		}
	}
}

func TestCommandTemplate_AllKnownWorkflows(t *testing.T) {
	knownWorkflows := []string{"propose", "explore", "apply", "archive", "new", "continue", "ff", "sync", "bulk-archive", "verify", "onboard"}
	for _, wf := range knownWorkflows {
		tmpl := CommandTemplate(wf)
		if tmpl.ID == "" {
			t.Errorf("CommandTemplate(%q): expected non-empty ID", wf)
		}
		if tmpl.Name == "" {
			t.Errorf("CommandTemplate(%q): expected non-empty Name", wf)
		}
		if tmpl.Body == "" {
			t.Errorf("CommandTemplate(%q): expected non-empty Body", wf)
		}
		if tmpl.Category == "" {
			t.Errorf("CommandTemplate(%q): expected non-empty Category", wf)
		}
		if len(tmpl.Tags) == 0 {
			t.Errorf("CommandTemplate(%q): expected non-empty Tags", wf)
		}
	}
}

func TestLoadSkillContent_KnownWorkflow(t *testing.T) {
	content := loadSkillContent("propose")
	if content == "" {
		t.Error("expected non-empty content for known workflow 'propose'")
	}
}

func TestLoadSkillContent_UnknownWorkflow(t *testing.T) {
	content := loadSkillContent("nonexistent-workflow-xyz")
	if content != "" {
		t.Errorf("expected empty content for unknown workflow, got %q", content)
	}
}

func TestLoadCommandContent_KnownWorkflow(t *testing.T) {
	content := loadCommandContent("propose")
	if content == "" {
		t.Error("expected non-empty content for known workflow 'propose'")
	}
}

func TestLoadCommandContent_UnknownWorkflow(t *testing.T) {
	content := loadCommandContent("nonexistent-workflow-xyz")
	if content != "" {
		t.Errorf("expected empty content for unknown workflow, got %q", content)
	}
}
