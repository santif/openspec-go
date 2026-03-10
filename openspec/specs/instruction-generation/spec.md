## ADDED Requirements

### Requirement: Instruction enrichment
The system SHALL load an artifact's instruction text from the active schema and enrich it by injecting project context (from `openspec/config.yaml` context field) and artifact-specific rules (from the rules map keyed by artifact ID). The enriched instruction MUST be a single text output ready for AI consumption.

#### Scenario: Enrich instruction with project context and rules
- **WHEN** `LoadEnrichedInstruction()` is called for the "specs" artifact and the project config has context text and a rule for "specs"
- **THEN** the returned instruction includes the schema's base instruction, the project context, and the specs-specific rule

#### Scenario: Enrich instruction with no project context
- **WHEN** `LoadEnrichedInstruction()` is called and the project config has no context or rules
- **THEN** the returned instruction contains only the schema's base instruction text

### Requirement: Apply instruction
The system SHALL provide special handling for the `apply` artifact instruction. The apply instruction MUST be enriched with the schema's apply phase instruction, project context, and rules — but does NOT include task progress. Task progress is handled separately by the status command.

#### Scenario: Apply instruction includes schema and context
- **WHEN** `LoadApplyInstruction()` is called for a change with project context configured
- **THEN** the returned instruction includes the apply phase instruction enriched with schema instruction and project context

#### Scenario: Apply instruction with no project context
- **WHEN** `LoadApplyInstruction()` is called and the project config has no context or rules
- **THEN** the returned instruction includes only the schema's apply phase instruction text

### Requirement: Artifact listing
The system SHALL list all available artifacts from the active schema when `openspec instructions` is run with no artifact argument.

#### Scenario: List available artifacts
- **WHEN** `openspec instructions` is run with no argument
- **THEN** the system lists all artifact IDs from the active schema (e.g., proposal, specs, design, tasks)

### Requirement: Change scoping
The system SHALL accept a `--change` flag that scopes the instruction to a specific change directory. When scoped, the system reads the change's metadata to determine the schema, providing schema-specific context for instruction enrichment.

#### Scenario: Scoped instruction for a change
- **WHEN** `openspec instructions specs --change my-change` is run
- **THEN** the instruction is enriched with context from the `my-change` directory

### Requirement: Output formats
The system SHALL output instructions as plain text by default and as JSON when `--json` is specified. The JSON format SHALL include fields for the artifact ID, instruction text, and any metadata.

#### Scenario: Plain text output
- **WHEN** `openspec instructions proposal --change my-change` is run without `--json`
- **THEN** the enriched instruction text is printed to stdout

#### Scenario: JSON output
- **WHEN** `openspec instructions proposal --change my-change --json` is run
- **THEN** the output is a JSON object with artifact and instruction fields
