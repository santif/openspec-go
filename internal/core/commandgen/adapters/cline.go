package adapters

import (
	"fmt"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// ClineAdapter generates commands for Cline with markdown headers instead of YAML.
type ClineAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Cline-specific command files in .clinerules/workflows/.
func (a *ClineAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatClineCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      ".clinerules/workflows",
		})
	}
	return results
}

// formatClineCommand formats a command with markdown headers (no YAML frontmatter).
func formatClineCommand(tmpl commandgen.CommandTemplateData) string {
	return fmt.Sprintf("# %s\n\n> %s\n\n%s",
		tmpl.Name,
		tmpl.Description,
		tmpl.Body,
	)
}
