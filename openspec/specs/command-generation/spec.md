## Purpose
Adapter-based system that generates skill and command markdown files for multiple AI tools, dispatching to per-tool adapters via a registry and supporting configurable delivery modes, workflow mapping, and tool detection.

## Requirements

### Requirement: Tool adapter system
The system SHALL maintain a registry of tool adapters, one per supported AI tool. Each adapter SHALL implement the `ToolCommandAdapter` interface providing `GetToolID()`, `GenerateSkills()`, `GenerateCommands()`, and `GetSkillsDir()`. The registry MUST support lookup by tool ID.

#### Scenario: Register and retrieve an adapter
- **WHEN** the Claude adapter is registered and `Get("claude")` is called
- **THEN** the registry returns the Claude adapter

#### Scenario: Lookup non-existent adapter
- **WHEN** `Get("unknown-tool")` is called
- **THEN** the registry returns nil or an error indicating no adapter found

### Requirement: Skill file generation
The system SHALL generate skill directory and files per workflow for a given AI tool. Each skill file SHALL contain the workflow-specific instruction content formatted for the target tool.

#### Scenario: Generate skills for Claude Code
- **WHEN** `GenerateSkills()` is called on the Claude adapter with workflows [propose, explore, apply, archive]
- **THEN** four skill files are generated with Claude-specific formatting in the `.claude/skills/` directory structure

#### Scenario: Generate skills with tool-specific formatting
- **WHEN** skill files are generated for Cursor
- **THEN** the files use Cursor-specific format (YAML frontmatter with appropriate fields)

### Requirement: Command file generation
The system SHALL generate command files per workflow for a given AI tool. Command files provide an alternative delivery mechanism for tools that support custom commands.

#### Scenario: Generate commands for a tool
- **WHEN** `GenerateCommands()` is called with workflows [propose, apply]
- **THEN** two command files are generated in the tool's command directory

### Requirement: Delivery modes
The system SHALL support three delivery modes: skills-only, commands-only, and both. The default delivery mode SHALL be "both". The delivery mode SHALL be read from global config and determines which file types are generated.

#### Scenario: Skills-only delivery
- **WHEN** delivery mode is "skills" and generation is triggered
- **THEN** only skill files are generated, no command files

#### Scenario: Both delivery mode
- **WHEN** delivery mode is "both" and generation is triggered
- **THEN** both skill files and command files are generated

### Requirement: Tool detection
The system SHALL scan the project directory for existing AI tool skill directories (`.claude/`, `.cursor/`, `.copilot/`, etc.) to determine which tools are currently installed.

#### Scenario: Detect installed Claude Code
- **WHEN** the project contains a `.claude/` directory
- **THEN** tool detection includes "claude" in the list of installed tools

#### Scenario: No tools detected
- **WHEN** the project contains no AI tool directories
- **THEN** tool detection returns an empty list

### Requirement: Update command
The system SHALL provide an `openspec update` command that detects installed AI tools and regenerates all skill/command files for those tools based on current configuration. A `--force` flag SHALL overwrite all files unconditionally.

#### Scenario: Update regenerates files for detected tools
- **WHEN** `openspec update` is run and Claude and Cursor are detected
- **THEN** skill/command files are regenerated for both tools

#### Scenario: Force update overwrites all files
- **WHEN** `openspec update --force` is run
- **THEN** all files are overwritten regardless of whether they have changed

### Requirement: Workflow-to-skill mapping
The system SHALL map 11 workflows to skill directory names: propose, explore, apply, archive, new, continue, ff, sync, bulk-archive, verify, and onboard. Each workflow name MUST map to a specific skill file name via the `WorkflowToSkillDir` configuration.

#### Scenario: Map workflow to skill name
- **WHEN** skill files are generated for the "propose" workflow
- **THEN** the generated file uses the skill name from `WorkflowToSkillDir["propose"]`

### Requirement: Specialized vs generic adapters
The system SHALL provide specialized adapters for 7 AI tools (Claude, Cursor, Codex, Cline, OpenCode, Factory, Windsurf) with custom formatting specific to each tool's conventions. All remaining supported tools SHALL use a generic BaseAdapter that produces a standard format. The adapter registry MUST map each tool ID to its appropriate adapter.

#### Scenario: Specialized adapter for Claude
- **WHEN** skill files are generated for Claude Code
- **THEN** the Claude-specific adapter is used with Claude's skill format conventions

#### Scenario: Generic adapter for unsupported tool
- **WHEN** skill files are generated for a tool without a specialized adapter
- **THEN** the generic BaseAdapter is used to produce standard-format skill files
