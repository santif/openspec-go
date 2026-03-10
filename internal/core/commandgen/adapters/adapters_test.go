package adapters

import (
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

var testWorkflows = []string{"propose"}

func TestBaseAdapter_GenerateSkills(t *testing.T) {
	adapter := &BaseAdapter{ToolID: "test-base", SkillsDir: ".testbase"}
	results := adapter.GenerateSkills(testWorkflows, "1.0.0")

	if len(results) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(results))
	}
	r := results[0]
	if r.FileName != "openspec-propose.md" {
		t.Errorf("FileName = %q, want %q", r.FileName, "openspec-propose.md")
	}
	if r.Dir != ".testbase/skills" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".testbase/skills")
	}
	if !strings.Contains(r.Content, "---") {
		t.Error("expected YAML frontmatter in skill content")
	}
	if !strings.Contains(r.Content, "name:") {
		t.Error("expected name field in skill content")
	}
}

func TestBaseAdapter_GenerateCommands(t *testing.T) {
	adapter := &BaseAdapter{ToolID: "test-base", SkillsDir: ".testbase"}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.FileName != "opsx-propose.md" {
		t.Errorf("FileName = %q, want %q", r.FileName, "opsx-propose.md")
	}
	if r.Dir != ".testbase/commands" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".testbase/commands")
	}
	if !strings.Contains(r.Content, "description:") {
		t.Error("expected description in command content")
	}
}

func TestClaudeAdapter_GenerateCommands(t *testing.T) {
	adapter := &ClaudeAdapter{BaseAdapter: BaseAdapter{ToolID: "claude", SkillsDir: ".claude"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.FileName != "propose.md" {
		t.Errorf("FileName = %q, want %q", r.FileName, "propose.md")
	}
	if r.Dir != ".claude/commands/opsx" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".claude/commands/opsx")
	}
	if !strings.Contains(r.Content, "name:") {
		t.Error("expected name field")
	}
	if !strings.Contains(r.Content, "category:") {
		t.Error("expected category field")
	}
	if !strings.Contains(r.Content, "tags:") {
		t.Error("expected tags field")
	}
}

func TestCursorAdapter_GenerateCommands(t *testing.T) {
	adapter := &CursorAdapter{BaseAdapter: BaseAdapter{ToolID: "cursor", SkillsDir: ".cursor"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.FileName != "opsx-propose.md" {
		t.Errorf("FileName = %q, want %q", r.FileName, "opsx-propose.md")
	}
	if r.Dir != ".cursor/commands" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".cursor/commands")
	}
	if !strings.Contains(r.Content, "name: /opsx-propose") {
		t.Error("expected Cursor-style name field")
	}
	if !strings.Contains(r.Content, "id: opsx-propose") {
		t.Error("expected id field")
	}
}

func TestClineAdapter_GenerateCommands(t *testing.T) {
	adapter := &ClineAdapter{BaseAdapter: BaseAdapter{ToolID: "cline", SkillsDir: ".cline"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.Dir != ".clinerules/workflows" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".clinerules/workflows")
	}
	// Cline uses markdown headers, not YAML frontmatter
	if strings.HasPrefix(r.Content, "---") {
		t.Error("Cline commands should NOT have YAML frontmatter")
	}
	if !strings.HasPrefix(r.Content, "# ") {
		t.Error("expected markdown header at start")
	}
	if !strings.Contains(r.Content, "> ") {
		t.Error("expected blockquote description")
	}
}

func TestCodexAdapter_GenerateCommands(t *testing.T) {
	t.Setenv("CODEX_HOME", "/tmp/test-codex")

	adapter := &CodexAdapter{BaseAdapter: BaseAdapter{ToolID: "codex", SkillsDir: ".codex"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.Dir != "/tmp/test-codex/prompts" {
		t.Errorf("Dir = %q, want %q", r.Dir, "/tmp/test-codex/prompts")
	}
	if !strings.Contains(r.Content, "argument-hint:") {
		t.Error("expected argument-hint field")
	}
}

func TestOpenCodeAdapter_GenerateSkills(t *testing.T) {
	adapter := &OpenCodeAdapter{BaseAdapter: BaseAdapter{ToolID: "opencode", SkillsDir: ".opencode"}}
	results := adapter.GenerateSkills(testWorkflows, "1.0.0")

	if len(results) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(results))
	}
	r := results[0]
	if r.Dir != ".opencode/skills" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".opencode/skills")
	}
	// OpenCode transforms /opsx: to /opsx-
	if strings.Contains(r.Content, "/opsx:") {
		t.Error("expected /opsx: to be transformed to /opsx-")
	}
}

func TestOpenCodeAdapter_GenerateCommands(t *testing.T) {
	adapter := &OpenCodeAdapter{BaseAdapter: BaseAdapter{ToolID: "opencode", SkillsDir: ".opencode"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.Dir != ".opencode/commands" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".opencode/commands")
	}
	if strings.Contains(r.Content, "/opsx:") {
		t.Error("expected /opsx: to be transformed to /opsx-")
	}
}

func TestFactoryAdapter_GenerateCommands(t *testing.T) {
	adapter := &FactoryAdapter{BaseAdapter: BaseAdapter{ToolID: "factory", SkillsDir: ".factory"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.Dir != ".factory/commands" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".factory/commands")
	}
	if !strings.Contains(r.Content, "argument-hint:") {
		t.Error("expected argument-hint field")
	}
}

func TestWindsurfAdapter_GenerateCommands(t *testing.T) {
	adapter := &WindsurfAdapter{BaseAdapter: BaseAdapter{ToolID: "windsurf", SkillsDir: ".windsurf"}}
	results := adapter.GenerateCommands(testWorkflows)

	if len(results) != 1 {
		t.Fatalf("expected 1 command, got %d", len(results))
	}
	r := results[0]
	if r.Dir != ".windsurf/workflows" {
		t.Errorf("Dir = %q, want %q", r.Dir, ".windsurf/workflows")
	}
	if !strings.Contains(r.Content, "name:") {
		t.Error("expected name field")
	}
	if !strings.Contains(r.Content, "category:") {
		t.Error("expected category field")
	}
	if !strings.Contains(r.Content, "tags:") {
		t.Error("expected tags field")
	}
}

func TestRegisteredAdapters(t *testing.T) {
	// Verify adapters were registered by init()
	all := commandgen.AllAdapters()
	expectedTools := []string{"claude", "cursor", "cline", "codex", "opencode", "factory", "windsurf"}
	for _, tool := range expectedTools {
		if all[tool] == nil {
			t.Errorf("expected adapter registered for %q", tool)
		}
	}
}

func TestBaseAdapter_MultipleWorkflows(t *testing.T) {
	adapter := &BaseAdapter{ToolID: "multi", SkillsDir: ".multi"}
	workflows := []string{"propose", "explore", "apply"}

	skills := adapter.GenerateSkills(workflows, "1.0.0")
	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}

	commands := adapter.GenerateCommands(workflows)
	if len(commands) != 3 {
		t.Errorf("expected 3 commands, got %d", len(commands))
	}
}
