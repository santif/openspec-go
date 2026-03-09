package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// FactoryAdapter generates commands for Factory with description + argument-hint.
type FactoryAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Factory-specific command files.
func (a *FactoryAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatFactoryCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(a.SkillsDir, "commands"),
		})
	}
	return results
}

// formatFactoryCommand formats a command with Factory's frontmatter:
// description, argument-hint.
func formatFactoryCommand(tmpl commandgen.CommandTemplateData) string {
	return fmt.Sprintf("---\ndescription: %s\nargument-hint: command arguments\n---\n\n%s",
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Body,
	)
}
