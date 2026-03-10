## ADDED Requirements

### Requirement: Spec validation
The system SHALL validate spec files by checking for required sections (Purpose, Requirements), verifying that requirements are at heading level 3 (`###`), verifying that scenarios are at heading level 4 (`####`), and checking for the presence of normative keywords (SHALL or MUST) in requirement text.

#### Scenario: Validate a well-formed spec
- **WHEN** a spec file contains Purpose, Requirements with `###` requirements each having `####` scenarios with SHALL statements
- **THEN** the validation report returns Valid=true with no error-level issues

#### Scenario: Validate a spec missing the Requirements section
- **WHEN** a spec file contains Purpose but no Requirements section
- **THEN** the validation report returns Valid=false with an error indicating the Requirements section is missing

#### Scenario: Validate a spec with scenarios at wrong heading level
- **WHEN** a spec has scenarios using `###` instead of `####`
- **THEN** the validation report includes a warning about incorrect scenario heading level

### Requirement: Change validation
The system SHALL validate change files by checking for required sections (Why, What Changes), enforcing Why section length between 50 and 1000 characters, verifying delta presence, and flagging an error when a change references more than 20 unique spec names.

#### Scenario: Validate a change with Why section too short
- **WHEN** a change file has a Why section with only 20 characters
- **THEN** the validation report returns Valid=false with an error indicating Why section is below the 50-character minimum

#### Scenario: Validate a change with too many unique specs
- **WHEN** a change file references 21 unique spec names in its delta bullets
- **THEN** the validation report returns an error indicating the change exceeds the 20 unique spec limit

### Requirement: Delta spec validation
The system SHALL validate delta spec files by checking for valid operation headers (ADDED, MODIFIED, REMOVED, RENAMED), verifying requirement block structure within each operation, and ensuring each requirement has at least one scenario.

#### Scenario: Validate a delta spec with invalid operation header
- **WHEN** a delta spec file contains `## UPDATED Requirements` instead of a valid operation header
- **THEN** the unrecognized header is silently ignored during parsing and no requirements are extracted from that section

#### Scenario: Validate a delta spec with a requirement missing scenarios
- **WHEN** a delta spec contains a requirement under ADDED with no `####` scenario subsections
- **THEN** the validation report returns an error indicating the requirement has no scenarios

### Requirement: Cross-section conflict detection
The system SHALL detect conflicts when the same requirement name appears in multiple operation sections within a single delta spec. A requirement MUST NOT appear in both MODIFIED and REMOVED sections, or any other conflicting combination.

#### Scenario: Detect requirement in both MODIFIED and REMOVED
- **WHEN** a delta spec has requirement "Auth" under both `## MODIFIED Requirements` and `## REMOVED Requirements`
- **THEN** the validation report returns an error indicating a conflict for "Auth"

### Requirement: Delta requirement content validation
The system SHALL validate requirements within ADDED and MODIFIED delta sections using the same rules as main spec requirements. Each requirement MUST contain normative keywords (SHALL or MUST) and MUST have at least one scenario.

#### Scenario: Delta requirement missing normative keyword
- **WHEN** a requirement under `## ADDED Requirements` in a delta spec lacks SHALL or MUST keywords
- **THEN** the validation report returns an error for the missing normative keyword

### Requirement: Validation report structure
The system SHALL produce validation reports containing a Valid boolean, a list of Issues (each with Level, Path, and Message), and a Summary with counts of errors, warnings, and info items. Issue levels SHALL be one of: error, warning, or info.

#### Scenario: Report with mixed issue levels
- **WHEN** validation finds 1 error and 2 warnings
- **THEN** the report has Valid=false, Issues contains 3 items, and Summary shows errors=1, warnings=2

#### Scenario: Report with no issues
- **WHEN** validation finds no issues
- **THEN** the report has Valid=true, Issues is empty, and Summary shows all counts at zero

### Requirement: Batch validation
The system SHALL support concurrent validation of multiple items using goroutines. Results from all items SHALL be merged into a single report. The system MUST use synchronization to safely aggregate concurrent results.

#### Scenario: Validate all changes concurrently
- **WHEN** `openspec validate --changes` is run with 5 active changes
- **THEN** all 5 changes are validated concurrently and results are merged into a single report

#### Scenario: Validate all items
- **WHEN** `openspec validate --all` is run
- **THEN** both specs and changes are validated and a combined report is produced

### Requirement: Strict mode
The system SHALL support a `--strict` flag that causes the validation report to be marked as invalid when any warning-level issues exist. Warnings SHALL remain at warning level but their presence SHALL cause `report.Valid` to become false.

#### Scenario: Warning becomes error in strict mode
- **WHEN** validation produces a warning and `--strict` is enabled
- **THEN** the issue remains at warning level but the report has Valid=false

### Requirement: CLI integration for validation
The system SHALL provide an `openspec validate` command that validates a single item by name, auto-detects whether the item is a change or spec, supports `--all`, `--changes`, and `--specs` flags for batch validation, and supports `--json` for JSON-formatted output.

#### Scenario: Validate a single change by name
- **WHEN** `openspec validate my-change` is run and `my-change` exists as a change
- **THEN** the system validates the change and displays the report

#### Scenario: Auto-detect item type
- **WHEN** `openspec validate user-auth` is run and `user-auth` exists as a spec but not a change
- **THEN** the system validates it as a spec

#### Scenario: JSON output format
- **WHEN** `openspec validate my-change --json` is run
- **THEN** the validation report is output as a JSON object with valid, issues, and summary fields
