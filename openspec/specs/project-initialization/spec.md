## Purpose
Project initialization command that creates the `openspec/` directory structure, generates `config.yaml`, configures AI tool integrations, supports force mode for legacy cleanup, and allows profile selection.

## Requirements

### Requirement: Directory structure creation
The system SHALL create the `openspec/` directory with `specs/` and `changes/` subdirectories when `openspec init` is run in a project root. Directories SHALL be created using `EnsureDir` which creates them if they do not already exist.

#### Scenario: Initialize a fresh project
- **WHEN** `openspec init` is run in a directory with no existing `openspec/` directory
- **THEN** the system creates `openspec/`, `openspec/specs/`, and `openspec/changes/` directories

#### Scenario: Directories already exist
- **WHEN** `openspec init` is run in a directory that already contains `openspec/specs/` and `openspec/changes/`
- **THEN** the system ensures the directories exist without error (idempotent directory creation)

### Requirement: Config file generation
The system SHALL write an `openspec/config.yaml` file during initialization only if the file does not already exist. The schema SHALL be hardcoded to `spec-driven` — no `--schema` flag exists.

#### Scenario: Generate config with default schema
- **WHEN** `openspec init` is run and no `openspec/config.yaml` exists
- **THEN** `openspec/config.yaml` is created with `schema: spec-driven`

#### Scenario: Skip config when already exists
- **WHEN** `openspec init` is run and `openspec/config.yaml` already exists
- **THEN** the existing config file is preserved and not overwritten

### Requirement: AI tool selection
The system SHALL accept a `--tools` flag for non-interactive tool selection. The flag takes a comma-separated list of tool IDs. No interactive prompt is implemented — tool selection is only available via the `--tools` flag.

#### Scenario: Non-interactive tool selection
- **WHEN** `openspec init --tools claude,cursor` is run
- **THEN** skill directories for Claude Code and Cursor are created with appropriate skill files

#### Scenario: Init without tools flag
- **WHEN** `openspec init` is run without `--tools`
- **THEN** no tool-specific skill files are generated

### Requirement: Force mode for legacy cleanup
The system SHALL support a `--force` flag that controls legacy cleanup behavior. The flag does NOT control overwrite behavior for configuration or directories — it only triggers legacy directory and marker block removal.

#### Scenario: Force triggers legacy cleanup
- **WHEN** `openspec init --force` is run on a project with legacy OpenSpec patterns
- **THEN** the system removes legacy slash-command directories and marker blocks before proceeding with initialization

### Requirement: Profile selection
The system SHALL accept a `--profile` flag to set the workflow profile during initialization. The flag value is stored but skill generation uses the global config profile for determining which workflows to generate.

#### Scenario: Initialize with profile flag
- **WHEN** `openspec init --profile core` is run
- **THEN** the profile value is read from the flag and stored in configuration
