package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// CursorAdapter generates commands for Cursor with its specific frontmatter format.
type CursorAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Cursor-specific command files.
func (a *CursorAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatCursorCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(a.SkillsDir, "commands"),
		})
	}
	return results
}

// formatCursorCommand formats a command with Cursor's YAML frontmatter:
// name as /opsx-<id>, id, category, description.
func formatCursorCommand(tmpl commandgen.CommandTemplateData) string {
	return fmt.Sprintf("---\nname: /opsx-%s\nid: opsx-%s\ncategory: %s\ndescription: %s\n---\n\n%s",
		tmpl.ID,
		tmpl.ID,
		tmpl.Category,
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Body,
	)
}
