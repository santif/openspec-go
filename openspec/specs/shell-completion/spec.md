## Purpose
Shell completion script generator that produces bash, zsh, fish, and powershell completions for all CLI commands and flags, with shell argument validation and stdout output.

## Requirements

### Requirement: Completion script generation
The system SHALL generate shell-specific completion scripts for bash, zsh, fish, and powershell when `openspec completion <shell>` is run. Each generated script MUST provide completions for all commands, subcommands, and flags.

#### Scenario: Generate bash completion
- **WHEN** `openspec completion bash` is run
- **THEN** a bash completion script is written to stdout

#### Scenario: Generate zsh completion
- **WHEN** `openspec completion zsh` is run
- **THEN** a zsh completion script is written to stdout

#### Scenario: Generate fish completion
- **WHEN** `openspec completion fish` is run
- **THEN** a fish completion script is written to stdout

#### Scenario: Generate powershell completion
- **WHEN** `openspec completion powershell` is run
- **THEN** a powershell completion script is written to stdout

### Requirement: Shell argument validation
The system SHALL accept exactly one argument from the valid set: bash, zsh, fish, powershell. The system MUST reject unsupported shell names with an error message listing valid options.

#### Scenario: Reject unsupported shell
- **WHEN** `openspec completion ksh` is run
- **THEN** the system returns an error listing the valid shell names

#### Scenario: Reject missing argument
- **WHEN** `openspec completion` is run with no shell argument
- **THEN** the system returns an error indicating a shell name is required

### Requirement: Stdout output
The system SHALL write completion scripts to stdout, allowing users to pipe or redirect the output to the appropriate shell configuration file.

#### Scenario: Pipe completion to file
- **WHEN** `openspec completion bash > ~/.bash_completion.d/openspec` is run
- **THEN** the completion script is written to the specified file via shell redirection
