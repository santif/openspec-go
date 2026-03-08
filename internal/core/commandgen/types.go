package commandgen

// CommandContent represents a generated file to be written to disk.
type CommandContent struct {
	FileName string
	Content  string
	Dir      string // relative directory path where the file should be written
}

// ToolCommandAdapter generates skills and commands for a specific AI tool.
type ToolCommandAdapter interface {
	GetToolID() string
	GenerateSkills(workflows []string) []CommandContent
	GenerateCommands(workflows []string) []CommandContent
	GetSkillsDir() string
}
