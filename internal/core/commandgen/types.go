package commandgen

// CommandContent represents a generated file to be written to disk.
type CommandContent struct {
	FileName string
	Content  string
	Dir      string // relative directory path where the file should be written
}

// SkillTemplateData holds the metadata and content for a workflow skill.
type SkillTemplateData struct {
	Name          string
	Description   string
	Instructions  string
	License       string
	Compatibility string
	Author        string
	Version       string
}

// CommandTemplateData holds the metadata and content for a workflow command.
type CommandTemplateData struct {
	ID          string
	Name        string
	Description string
	Category    string
	Tags        []string
	Body        string
}

// ToolCommandAdapter generates skills and commands for a specific AI tool.
type ToolCommandAdapter interface {
	GetToolID() string
	GenerateSkills(workflows []string, version string) []CommandContent
	GenerateCommands(workflows []string) []CommandContent
	GetSkillsDir() string
}
