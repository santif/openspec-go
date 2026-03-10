## Purpose
Schema definition and resolution system that loads YAML-based artifact schemas with dependency validation, supports a three-tier resolution chain (user override, project-local, built-in embedded), and provides forking, listing, and inspection capabilities.

## Requirements

### Requirement: Schema YAML structure
The system SHALL define schemas as YAML files containing a name, version, description, and a list of artifacts. Each artifact SHALL have an id, description (text), generates (glob pattern), template (file path), instruction (text), and requires (list of artifact IDs). An optional apply phase MAY be defined with its own requires and tracks fields.

#### Scenario: Parse a valid schema with four artifacts
- **WHEN** a schema YAML defines artifacts: proposal, specs, design, tasks with dependency chain proposal→specs→design→tasks
- **THEN** `ParseSchema()` returns a SchemaYaml with four artifacts, each with their respective generates patterns and dependency lists

#### Scenario: Parse a schema with an apply phase
- **WHEN** a schema YAML includes an `apply` section with `requires: [tasks]` and `tracks: tasks.md`
- **THEN** the parsed schema includes the ApplyPhase with the correct requires and tracks values

### Requirement: Schema loading and validation
The system SHALL load schema YAML files and validate them for correctness. Validation MUST detect duplicate artifact IDs, invalid references in requires lists (referencing non-existent artifact IDs), and dependency cycles via depth-first search.

#### Scenario: Reject a schema with duplicate artifact IDs
- **WHEN** a schema YAML contains two artifacts both with id "proposal"
- **THEN** `LoadSchema()` returns an error indicating duplicate artifact ID

#### Scenario: Reject a schema with invalid reference
- **WHEN** a schema YAML has an artifact with `requires: [nonexistent]` and no artifact with id "nonexistent" exists
- **THEN** `LoadSchema()` returns an error indicating invalid reference

#### Scenario: Reject a schema with dependency cycle
- **WHEN** artifact A requires B and artifact B requires A
- **THEN** `LoadSchema()` returns an error indicating a dependency cycle was detected

### Requirement: Three-tier resolution chain
The system SHALL resolve schemas using a three-tier chain: user override (XDG data directory via `globalconfig.GetGlobalDataDir()`, typically `~/.local/share/openspec/schemas/`) takes highest precedence, followed by project-local (`openspec/schemas/`), followed by built-in embedded schemas. The first tier containing the requested schema wins.

#### Scenario: Project-local schema overrides built-in
- **WHEN** a project has `openspec/schemas/custom/schema.yaml` and no user override exists
- **THEN** `ResolveSchema("custom")` returns the project-local schema

#### Scenario: Built-in schema used as fallback
- **WHEN** no user override or project-local schema exists for "spec-driven"
- **THEN** `ResolveSchema("spec-driven")` returns the embedded built-in schema

#### Scenario: User override takes highest precedence
- **WHEN** a schema named "spec-driven" exists in all three tiers
- **THEN** `ResolveSchema("spec-driven")` returns the user override version

### Requirement: Embedded schemas
The system SHALL embed built-in schemas using Go's `//go:embed` directive. The default schema SHALL be named `spec-driven` and define four artifacts: proposal, specs, design, and tasks.

#### Scenario: Access the default built-in schema
- **WHEN** the system loads schemas with no project or user overrides
- **THEN** the `spec-driven` schema is available with proposal, specs, design, and tasks artifacts

### Requirement: Schema version validation
The system SHALL validate that the schema version field is a positive integer. `ParseSchema()` MUST return an error when the version is zero or negative. Semantic versioning is NOT used — versions are simple incrementing integers.

#### Scenario: Reject schema with zero version
- **WHEN** a schema YAML has `version: 0`
- **THEN** `ParseSchema()` returns an error indicating the version must be positive

#### Scenario: Accept schema with valid version
- **WHEN** a schema YAML has `version: 1`
- **THEN** `ParseSchema()` succeeds and returns the schema with version 1

### Requirement: Schema forking
The system SHALL support a `schema fork` command that copies a built-in or resolved schema and its templates to the project's `openspec/schemas/` directory, enabling local customization.

#### Scenario: Fork the default schema
- **WHEN** `openspec schema fork spec-driven` is run
- **THEN** a minimal schema.yaml (containing only name and version) and all template files are created in `openspec/schemas/spec-driven/`

#### Scenario: Fork when local copy already exists
- **WHEN** `openspec schema fork spec-driven` is run and `openspec/schemas/spec-driven/` already exists
- **THEN** the system warns that the schema already exists locally

### Requirement: Schema listing
The system SHALL provide a command to list all available schemas across all resolution tiers, annotating each with its source (user, project, built-in). The listing SHALL support JSON output format.

#### Scenario: List schemas with source annotation
- **WHEN** `openspec schemas` is run and there is a built-in "spec-driven" and a project-local "custom" schema
- **THEN** the output lists both schemas with their respective source annotations

#### Scenario: JSON output for schema listing
- **WHEN** `openspec schemas --json` is run
- **THEN** the output is a JSON array of schema objects with name and source fields

### Requirement: Schema inspection
The system SHALL provide `schema which` to show the active schema for the current project and `schema validate` to check a schema file's integrity.

#### Scenario: Show active schema
- **WHEN** `openspec schema which` is run in a project configured with schema "spec-driven"
- **THEN** the output shows only the schema name "spec-driven"

#### Scenario: Validate schema integrity
- **WHEN** `openspec schema validate` is run on a project with a valid schema
- **THEN** the output confirms the schema is valid with no errors
