package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// BaseAdapter provides shared skill/command generation for tools that use
// markdown skill files in a skills/ subdirectory.
type BaseAdapter struct {
	ToolID    string
	SkillsDir string // e.g. ".claude"
}

func (b *BaseAdapter) GetToolID() string {
	return b.ToolID
}

func (b *BaseAdapter) GetSkillsDir() string {
	return b.SkillsDir
}

// GenerateSkills produces markdown skill files with YAML frontmatter for each workflow.
func (b *BaseAdapter) GenerateSkills(workflows []string, version string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.SkillTemplate(wf)
		content := commandgen.GenerateSkillContent(tmpl, version)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("openspec-%s.md", wf),
			Content:  content,
			Dir:      filepath.Join(b.SkillsDir, "skills"),
		})
	}
	return results
}

// GenerateCommands produces command wrapper files for each workflow.
// The base adapter uses a simple frontmatter with description only.
func (b *BaseAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		content := formatDefaultCommand(tmpl)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(b.SkillsDir, "commands"),
		})
	}
	return results
}

// formatDefaultCommand formats a command with a simple description-only frontmatter.
func formatDefaultCommand(tmpl commandgen.CommandTemplateData) string {
	return fmt.Sprintf("---\ndescription: %s\n---\n\n%s",
		commandgen.EscapeYamlValue(tmpl.Description),
		tmpl.Body,
	)
}
