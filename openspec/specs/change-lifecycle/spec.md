## Purpose
Change management commands that handle creating, listing, displaying, archiving, and discovering change proposals within the `openspec/changes/` directory, including metadata tracking and spec listing.

## Requirements

### Requirement: Change creation
The system SHALL create a new change directory at `openspec/changes/<name>/` when `openspec new change <name>` is run. The name MUST be validated as kebab-case: it must start with a letter and match the pattern `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`. The system SHALL write a `proposal.md` from the schema template and a `.openspec.yaml` metadata file with the schema name and creation date.

#### Scenario: Create a valid change
- **WHEN** `openspec new change add-auth` is run
- **THEN** directory `openspec/changes/add-auth/` is created with `proposal.md` and `.openspec.yaml`

#### Scenario: Reject invalid change name
- **WHEN** `openspec new change "My Change"` is run
- **THEN** the system returns an error indicating the name must be kebab-case

#### Scenario: Reject duplicate change name
- **WHEN** `openspec new change existing-change` is run and that change already exists
- **THEN** the system returns an error indicating the change already exists

### Requirement: Change listing
The system SHALL list all active changes when `openspec list` is run. Changes SHALL be sorted by modification time (most recent first) by default. Alphabetical sorting SHALL be available as an option. JSON output format SHALL be supported.

#### Scenario: List changes sorted by time
- **WHEN** `openspec list` is run with 3 active changes
- **THEN** all 3 changes are listed with the most recently modified first

#### Scenario: List changes alphabetically
- **WHEN** `openspec list --sort name` is run
- **THEN** changes are listed in alphabetical order by name

#### Scenario: JSON output for listing
- **WHEN** `openspec list --json` is run
- **THEN** the output is a JSON array of change objects

### Requirement: Spec listing
The system SHALL list all specs when `openspec list --specs` is run, using the same sorting and output format options as change listing.

#### Scenario: List all specs
- **WHEN** `openspec list --specs` is run with 5 specs in `openspec/specs/`
- **THEN** all 5 specs are listed

### Requirement: Item display
The system SHALL display the full content of a change or spec when `openspec show <name>` is run. The system MUST auto-detect whether the name refers to a change or spec. JSON output SHALL be supported via `--json`. The `--deltas-only` flag SHALL filter the displayed content to show only delta information.

#### Scenario: Show a change with auto-detection
- **WHEN** `openspec show my-change` is run and `my-change` exists as a change
- **THEN** the full change content (proposal, deltas) is displayed

#### Scenario: Show with JSON output
- **WHEN** `openspec show my-change --json` is run
- **THEN** the change is displayed as a structured JSON object

#### Scenario: Show with deltas-only filter
- **WHEN** `openspec show my-change --deltas-only` is run
- **THEN** only the delta information is displayed, omitting the Why and other sections

### Requirement: Change archiving
The system SHALL archive a completed change when `openspec archive <name>` is run. The process MUST pre-validate the change, apply delta specs to main specs, and move the change directory to `openspec/archive/` with a `YYYY-MM-DD-` date prefix. Non-interactive mode SHALL be available via `-y`/`--yes`.

#### Scenario: Archive a valid change
- **WHEN** `openspec archive my-change` is run and the change passes validation
- **THEN** delta specs are applied, and the change directory moves to `openspec/archive/2026-03-09-my-change/`

#### Scenario: Archive with validation failure
- **WHEN** `openspec archive my-change` is run and the change fails validation
- **THEN** the system reports the validation errors and does not archive

#### Scenario: Skip spec application
- **WHEN** `openspec archive my-change --skip-specs` is run
- **THEN** the change is archived without applying delta specs to main specs

#### Scenario: Non-interactive archive
- **WHEN** `openspec archive my-change -y` is run
- **THEN** the system skips the confirmation prompt and proceeds directly

### Requirement: Item discovery
The system SHALL scan the filesystem to discover active changes in `openspec/changes/`, specs in `openspec/specs/`, and archived changes in `openspec/archive/`. Discovery SHALL return lists of IDs (directory names).

#### Scenario: Discover active changes
- **WHEN** `GetActiveChangeIDs()` is called and `openspec/changes/` contains directories `auth` and `payments`
- **THEN** the result includes ["auth", "payments"]

#### Scenario: Discover specs
- **WHEN** `GetSpecIDs()` is called and `openspec/specs/` contains directories `user-auth` and `data-export`
- **THEN** the result includes ["user-auth", "data-export"]

### Requirement: Change metadata
The system SHALL read and write `.openspec.yaml` metadata files per change directory. The metadata SHALL contain the schema name and creation date.

#### Scenario: Read change metadata
- **WHEN** `ReadChangeMetadata()` is called on a change with `.openspec.yaml` containing `schema: spec-driven` and `created: "2026-03-09"`
- **THEN** the result has Schema="spec-driven" and Created="2026-03-09"

#### Scenario: Write change metadata
- **WHEN** `WriteChangeMetadata()` is called with schema="spec-driven" and created="2026-03-09"
- **THEN** a `.openspec.yaml` file is written with those values
