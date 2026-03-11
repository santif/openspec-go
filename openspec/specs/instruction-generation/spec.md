## Purpose
Instruction enrichment system that loads artifact templates from the active schema and injects project context, rules, change scoping, and artifact listing to produce ready-to-use AI instructions in text or JSON format.
## Requirements
### Requirement: Instruction enrichment
The system SHALL load an artifact's instruction text from the active schema and enrich it by injecting project context (from `openspec/config.yaml` context field), artifact-specific rules (from the rules map keyed by artifact ID), and custom conditional keywords instruction (when configured). The enriched instruction MUST be a single text output ready for AI consumption. When the project has custom conditional keywords configured, the system SHALL append a "Project Keywords" instruction block to the enriched instruction.

#### Scenario: Enrich instruction with project context and rules
- **WHEN** `LoadEnrichedInstruction()` is called for the "specs" artifact and the project config has context text and a rule for "specs"
- **THEN** the returned instruction includes the schema's base instruction, the project context, and the specs-specific rule

#### Scenario: Enrich instruction with no project context
- **WHEN** `LoadEnrichedInstruction()` is called and the project config has no context or rules
- **THEN** the returned instruction contains only the schema's base instruction text

#### Scenario: Enrich instruction with custom conditional keywords
- **WHEN** `LoadEnrichedInstruction()` is called and the project config has `keywords: { conditionals: { when: "CUANDO", then: "ENTONCES", and: "Y" } }`
- **THEN** the returned instruction includes a "Project Keywords" block instructing the AI to use CUANDO/ENTONCES/Y instead of WHEN/THEN/AND

#### Scenario: Enrich instruction without custom conditional keywords
- **WHEN** `LoadEnrichedInstruction()` is called and the project config has no conditionals configured
- **THEN** the returned instruction does not include a "Project Keywords" block (unchanged behavior)

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

