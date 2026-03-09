package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// ClaudeAdapter generates commands for Claude Code with its specific frontmatter format.
type ClaudeAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Claude-specific command files in .claude/commands/opsx/.
func (a *ClaudeAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatClaudeCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(a.SkillsDir, "commands", "opsx"),
		})
	}
	return results
}

// formatClaudeCommand formats a command with Claude's YAML frontmatter:
// name, description, category, tags.
func formatClaudeCommand(tmpl commandgen.CommandTemplateData) string {
	tags := commandgen.FormatTagsArray(tmpl.Tags)
	return fmt.Sprintf("---\nname: %s\ndescription: %s\ncategory: %s\ntags:%s\n---\n\n%s",
		commandgen.EscapeYamlValue(tmpl.Name),
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Category,
		tags,
		tmpl.Body,
	)
}
