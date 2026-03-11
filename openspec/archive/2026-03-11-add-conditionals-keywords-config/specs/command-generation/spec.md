## MODIFIED Requirements

### Requirement: Skill file generation
The system SHALL generate skill directory and files per workflow for a given AI tool. Each skill file SHALL contain the workflow-specific instruction content formatted for the target tool. When the project has custom conditional keywords configured, the system SHALL append a compact "Project Keywords" instruction block to the skill content instructing the AI to use the configured keywords instead of WHEN/THEN/AND.

#### Scenario: Generate skills for Claude Code
- **WHEN** `GenerateSkills()` is called on the Claude adapter with workflows [propose, explore, apply, archive]
- **THEN** four skill files are generated with Claude-specific formatting in the `.claude/skills/` directory structure

#### Scenario: Generate skills with tool-specific formatting
- **WHEN** skill files are generated for Cursor
- **THEN** the files use Cursor-specific format (YAML frontmatter with appropriate fields)

#### Scenario: Generate skills with custom conditional keywords
- **WHEN** skill files are generated and the project config has `keywords: { conditionals: { when: "CUANDO", then: "ENTONCES", and: "Y" } }`
- **THEN** each generated skill file includes a "Project Keywords" block instructing the AI to use CUANDO/ENTONCES/Y

#### Scenario: Generate skills without custom conditional keywords
- **WHEN** skill files are generated and the project config has no conditionals configured
- **THEN** skill files are generated without any "Project Keywords" block (unchanged behavior)

### Requirement: Command file generation
The system SHALL generate command files per workflow for a given AI tool. Command files provide an alternative delivery mechanism for tools that support custom commands. When the project has custom conditional keywords configured, the system SHALL append the same "Project Keywords" instruction block to command file content.

#### Scenario: Generate commands for a tool
- **WHEN** `GenerateCommands()` is called with workflows [propose, apply]
- **THEN** two command files are generated in the tool's command directory

#### Scenario: Generate commands with custom conditional keywords
- **WHEN** command files are generated and the project config has custom conditionals configured
- **THEN** each generated command file includes a "Project Keywords" block
