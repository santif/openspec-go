## MODIFIED Requirements

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
