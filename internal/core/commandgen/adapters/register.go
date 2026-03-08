package adapters

import "github.com/santif/openspec-go/internal/core/commandgen"

func init() {
	// Register all adapters. Each uses the BaseAdapter since they share
	// the same markdown skill file pattern, differing only in skills directory.
	tools := []struct {
		id        string
		skillsDir string
	}{
		{"amazon-q", ".amazonq"},
		{"antigravity", ".agent"},
		{"auggie", ".augment"},
		{"claude", ".claude"},
		{"cline", ".cline"},
		{"codex", ".codex"},
		{"codebuddy", ".codebuddy"},
		{"continue", ".continue"},
		{"costrict", ".cospec"},
		{"crush", ".crush"},
		{"cursor", ".cursor"},
		{"factory", ".factory"},
		{"gemini", ".gemini"},
		{"github-copilot", ".github"},
		{"iflow", ".iflow"},
		{"kilocode", ".kilocode"},
		{"kiro", ".kiro"},
		{"opencode", ".opencode"},
		{"pi", ".pi"},
		{"qoder", ".qoder"},
		{"qwen", ".qwen"},
		{"roocode", ".roo"},
		{"trae", ".trae"},
		{"windsurf", ".windsurf"},
	}

	for _, t := range tools {
		commandgen.Register(&BaseAdapter{
			ToolID:    t.id,
			SkillsDir: t.skillsDir,
		})
	}
}
