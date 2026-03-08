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

// GenerateSkills produces markdown skill files for each workflow.
func (b *BaseAdapter) GenerateSkills(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		content := commandgen.SkillTemplate(wf)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("openspec-%s.md", wf),
			Content:  content,
			Dir:      filepath.Join(b.SkillsDir, "skills"),
		})
	}
	return results
}

// GenerateCommands produces command wrapper files for each workflow.
// Most tools use the same skill files as commands, so this delegates to GenerateSkills.
func (b *BaseAdapter) GenerateCommands(workflows []string) []commandgen.CommandContent {
	var results []commandgen.CommandContent
	for _, wf := range workflows {
		content := commandgen.SkillTemplate(wf)
		results = append(results, commandgen.CommandContent{
			FileName: fmt.Sprintf("openspec-%s.md", wf),
			Content:  content,
			Dir:      filepath.Join(b.SkillsDir, "commands"),
		})
	}
	return results
}
