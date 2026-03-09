package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/santif/openspec-go/internal/core/commandgen"
)

// OpenCodeAdapter generates commands for OpenCode, transforming command references.
type OpenCodeAdapter struct {
	BaseAdapter
}

// GenerateSkills produces OpenCode skill files with /opsx: → /opsx- transforms.
func (a *OpenCodeAdapter) GenerateSkills(workflows []string, version string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.SkillTemplate(wf)
		// Transform colon-based command references to hyphen-based
		tmpl.Instructions = commandgen.TransformToHyphenCommands(tmpl.Instructions)
		content := commandgen.GenerateSkillContent(tmpl, version)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("openspec-%s.md", wf),
			Content:  content,
			Dir:      filepath.Join(a.SkillsDir, "skills"),
		})
	}
	return results
}

// GenerateCommands produces OpenCode-specific command files with transforms.
func (a *OpenCodeAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		tmpl := commandgen.CommandTemplate(wf)
		// Transform colon-based command references to hyphen-based
		body := commandgen.TransformToHyphenCommands(tmpl.Body)
		content := fmt.Sprintf("---\ndescription: %s\n---\n\n%s",
			commandgen.EscapeYamlValue(tmpl.Description),
			body,
		)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("opsx-%s.md", tmpl.ID),
			Content:  content,
			Dir:      filepath.Join(a.SkillsDir, "commands"),
		})
	}
	return results
}
