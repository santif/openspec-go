package adapters

import "github.com/santif/openspec-go/internal/core/commandgen"

func init() {
	// Tool-specific adapters with custom command formatting.
	commandgen.Register(&ClaudeAdapter{
		BaseAdapter: BaseAdapter{ToolID: "claude", SkillsDir: ".claude"},
	})
	commandgen.Register(&CursorAdapter{
		BaseAdapter: BaseAdapter{ToolID: "cursor", SkillsDir: ".cursor"},
	})
	commandgen.Register(&CodexAdapter{
		BaseAdapter: BaseAdapter{ToolID: "codex", SkillsDir: ".codex"},
	})
	commandgen.Register(&ClineAdapter{
		BaseAdapter: BaseAdapter{ToolID: "cline", SkillsDir: ".cline"},
	})
	commandgen.Register(&OpenCodeAdapter{
		BaseAdapter: BaseAdapter{ToolID: "opencode", SkillsDir: ".opencode"},
	})
	commandgen.Register(&FactoryAdapter{
		BaseAdapter: BaseAdapter{ToolID: "factory", SkillsDir: ".factory"},
	})
	commandgen.Register(&WindsurfAdapter{
		BaseAdapter: BaseAdapter{ToolID: "windsurf", SkillsDir: ".windsurf"},
	})

	// Generic adapters using BaseAdapter for all other tools.
	genericTools := []struct {
		id        string
		skillsDir string
	}{
		{"amazon-q", ".amazonq"},
		{"antigravity", ".agent"},
		{"auggie", ".augment"},
		{"codebuddy", ".codebuddy"},
		{"continue", ".continue"},
		{"costrict", ".cospec"},
		{"crush", ".crush"},
		{"gemini", ".gemini"},
		{"github-copilot", ".github"},
		{"iflow", ".iflow"},
		{"kilocode", ".kilocode"},
		{"kiro", ".kiro"},
		{"pi", ".pi"},
		{"qoder", ".qoder"},
		{"qwen", ".qwen"},
		{"roocode", ".roo"},
		{"trae", ".trae"},
	}

	for _, t := range genericTools {
		commandgen.Register(&BaseAdapter{
			ToolID:    t.id,
			SkillsDir: t.skillsDir,
		})
	}
}
