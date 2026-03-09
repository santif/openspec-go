package adapters

import (
	"fmt"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// WindsurfAdapter generates commands for Windsurf with its specific frontmatter.
type WindsurfAdapter struct {
	BaseAdapter
}

// GenerateCommands produces Windsurf-specific command files in .windsurf/workflows/.
func (a *WindsurfAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatWindsurfCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      ".windsurf/workflows",
		})
	}
	return results
}

// formatWindsurfCommand formats a command with Windsurf's YAML frontmatter:
// name, description, category, tags.
func formatWindsurfCommand(tmpl commandgen.CommandTemplateData) string {
	tags := commandgen.FormatTagsArray(tmpl.Tags)
	return fmt.Sprintf("---\nname: %s\ndescription: %s\ncategory: %s\ntags:%s\n---\n\n%s",
		commandgen.EscapeYamlValue(tmpl.Name),
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Category,
		tags,
		tmpl.Body,
	)
}
