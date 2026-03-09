package config

const OpenSpecDirName = "openspec"

var OpenSpecMarkers = struct {
	Start string
	End   string
}{
	Start: "<!-- OPENSPEC:START -->",
	End:   "<!-- OPENSPEC:END -->",
}

type AIToolOption struct {
	Name         string
	Value        string
	Available    bool
	SuccessLabel string
	SkillsDir    string // e.g., ".claude" - /skills suffix per Agent Skills spec
}

var AITools = []AIToolOption{
	{Name: "Amazon Q Developer", Value: "amazon-q", Available: true, SuccessLabel: "Amazon Q Developer", SkillsDir: ".amazonq"},
	{Name: "Antigravity", Value: "antigravity", Available: true, SuccessLabel: "Antigravity", SkillsDir: ".agent"},
	{Name: "Auggie (Augment CLI)", Value: "auggie", Available: true, SuccessLabel: "Auggie", SkillsDir: ".augment"},
	{Name: "Claude Code", Value: "claude", Available: true, SuccessLabel: "Claude Code", SkillsDir: ".claude"},
	{Name: "Cline", Value: "cline", Available: true, SuccessLabel: "Cline", SkillsDir: ".cline"},
	{Name: "Codex", Value: "codex", Available: true, SuccessLabel: "Codex", SkillsDir: ".codex"},
	{Name: "CodeBuddy Code (CLI)", Value: "codebuddy", Available: true, SuccessLabel: "CodeBuddy Code", SkillsDir: ".codebuddy"},
	{Name: "Continue", Value: "continue", Available: true, SuccessLabel: "Continue (VS Code / JetBrains / Cli)", SkillsDir: ".continue"},
	{Name: "CoStrict", Value: "costrict", Available: true, SuccessLabel: "CoStrict", SkillsDir: ".cospec"},
	{Name: "Crush", Value: "crush", Available: true, SuccessLabel: "Crush", SkillsDir: ".crush"},
	{Name: "Cursor", Value: "cursor", Available: true, SuccessLabel: "Cursor", SkillsDir: ".cursor"},
	{Name: "Factory Droid", Value: "factory", Available: true, SuccessLabel: "Factory Droid", SkillsDir: ".factory"},
	{Name: "Gemini CLI", Value: "gemini", Available: true, SuccessLabel: "Gemini CLI", SkillsDir: ".gemini"},
	{Name: "GitHub Copilot", Value: "github-copilot", Available: true, SuccessLabel: "GitHub Copilot", SkillsDir: ".github"},
	{Name: "iFlow", Value: "iflow", Available: true, SuccessLabel: "iFlow", SkillsDir: ".iflow"},
	{Name: "Kilo Code", Value: "kilocode", Available: true, SuccessLabel: "Kilo Code", SkillsDir: ".kilocode"},
	{Name: "Kiro", Value: "kiro", Available: true, SuccessLabel: "Kiro", SkillsDir: ".kiro"},
	{Name: "OpenCode", Value: "opencode", Available: true, SuccessLabel: "OpenCode", SkillsDir: ".opencode"},
	{Name: "Pi", Value: "pi", Available: true, SuccessLabel: "Pi", SkillsDir: ".pi"},
	{Name: "Qoder", Value: "qoder", Available: true, SuccessLabel: "Qoder", SkillsDir: ".qoder"},
	{Name: "Qwen Code", Value: "qwen", Available: true, SuccessLabel: "Qwen Code", SkillsDir: ".qwen"},
	{Name: "RooCode", Value: "roocode", Available: true, SuccessLabel: "RooCode", SkillsDir: ".roo"},
	{Name: "Trae", Value: "trae", Available: true, SuccessLabel: "Trae", SkillsDir: ".trae"},
	{Name: "Windsurf", Value: "windsurf", Available: true, SuccessLabel: "Windsurf", SkillsDir: ".windsurf"},
	{Name: "AGENTS.md (works with Amp, VS Code, ...)", Value: "agents", Available: false, SuccessLabel: "your AGENTS.md-compatible assistant", SkillsDir: ""},
}

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
