## Purpose
CLI command that displays resolved template file paths for all artifacts in the active or specified schema, with optional JSON output format.

## Requirements

### Requirement: Template path display
The system SHALL provide an `openspec templates` command that shows the resolved template file paths for all artifacts in the active schema. Each artifact's template path MUST be resolved through the schema resolution chain before display.

#### Scenario: Display template paths for default schema
- **WHEN** `openspec templates` is run in a project using the "spec-driven" schema
- **THEN** the output lists each artifact ID alongside its resolved template file path

#### Scenario: Display when no templates exist
- **WHEN** `openspec templates` is run and the schema has artifacts with no template files
- **THEN** the output indicates which artifacts have no template path

### Requirement: Schema flag
The system SHALL accept a `--schema` flag to inspect templates for a specific schema rather than the project's active schema. The specified schema MUST be resolvable through the schema resolution chain.

#### Scenario: Inspect templates for a specific schema
- **WHEN** `openspec templates --schema custom-workflow` is run
- **THEN** the output shows template paths resolved for the "custom-workflow" schema

### Requirement: JSON output
The system SHALL support a `--json` flag for machine-readable output. The JSON format SHALL include an array of objects with artifact ID and template path fields.

#### Scenario: JSON output for templates
- **WHEN** `openspec templates --json` is run
- **THEN** the output is a JSON array of objects each containing artifact and template fields
