package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// CodexAdapter generates commands for Codex, which uses global paths.
type CodexAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Codex-specific command files at the global codex directory.
func (a *CodexAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			codexHome = filepath.Join("~", ".codex")
		} else {
			codexHome = filepath.Join(home, ".codex")
		}
	}

	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatCodexCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(codexHome, "prompts"),
		})
	}
	return results
}

// formatCodexCommand formats a command with Codex's frontmatter:
// description, argument-hint.
func formatCodexCommand(tmpl commandgen.CommandTemplateData) string {
	return fmt.Sprintf("---\ndescription: %s\nargument-hint: command arguments\n---\n\n%s",
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Body,
	)
}
