package commandgen

import "embed"

//go:embed templates/skills/*.md templates/commands/*.md
var templateFS embed.FS

// loadSkillContent reads the embedded skill template for the given workflow.
func loadSkillContent(workflow string) string {
	data, err := templateFS.ReadFile("templates/skills/" + workflow + ".md")
	if err != nil {
		return ""
	}
	return string(data)
}

// loadCommandContent reads the embedded command template for the given workflow.
func loadCommandContent(workflow string) string {
	data, err := templateFS.ReadFile("templates/commands/" + workflow + ".md")
	if err != nil {
		return ""
	}
	return string(data)
}
