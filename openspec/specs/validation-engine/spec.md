## Purpose
Validation engine that checks specs and changes for structural correctness, normative keyword presence, section requirements, delta spec integrity, and cross-section conflicts, producing detailed reports with error/warning/info levels.
## Requirements
### Requirement: Spec validation
The system SHALL validate spec files by checking for required sections (Purpose, Requirements), verifying that requirements are at heading level 3 (`###`), verifying that scenarios are at heading level 4 (`####`), and checking for the presence of configured normative keywords in requirement text. When no custom keywords are configured, the system SHALL default to requiring SHALL or MUST. Validation guide messages (scenario format hints, missing section examples) SHALL use the project's configured conditional keywords instead of hardcoded WHEN/THEN/AND.

#### Scenario: Validate a well-formed spec
- **WHEN** a spec file contains Purpose, Requirements with `###` requirements each having `####` scenarios with SHALL statements
- **THEN** the validation report returns Valid=true with no error-level issues

#### Scenario: Validate a spec missing the Requirements section
- **WHEN** a spec file contains Purpose but no Requirements section
- **THEN** the validation report returns Valid=false with an error indicating the Requirements section is missing

#### Scenario: Validate a spec with scenarios at wrong heading level
- **WHEN** a spec has scenarios using `###` instead of `####`
- **THEN** the validation report includes a warning about incorrect scenario heading level

#### Scenario: Validate a spec with custom normative keywords
- **WHEN** the project config has `keywords: { normative: ["DEBE", "DEBERA"] }` and a spec requirement contains "DEBE" but not "SHALL" or "MUST"
- **THEN** the validation report returns Valid=true for that requirement

#### Scenario: Validate a spec failing custom normative keywords
- **WHEN** the project config has `keywords: { normative: ["DEBE", "DEBERA"] }` and a spec requirement contains neither "DEBE" nor "DEBERA"
- **THEN** the validation report returns an error indicating the requirement must contain DEBE or DEBERA keyword

#### Scenario: Guide message uses custom conditional keywords
- **WHEN** the project config has `keywords: { conditionals: { when: "CUANDO", then: "ENTONCES", and: "Y" } }` and a validation guide message is generated
- **THEN** the guide message shows `**CUANDO**`, `**ENTONCES**`, and `**Y**` instead of WHEN, THEN, and AND

#### Scenario: Guide message uses default conditional keywords
- **WHEN** the project config has no conditionals configured and a validation guide message is generated
- **THEN** the guide message shows `**WHEN**`, `**THEN**`, and `**AND**` (default behavior)

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
The system SHALL validate requirements within ADDED and MODIFIED delta sections using the same rules as main spec requirements. Each requirement MUST contain the configured normative keywords and MUST have at least one scenario. When no custom keywords are configured, the system SHALL default to requiring SHALL or MUST.

#### Scenario: Delta requirement missing normative keyword
- **WHEN** a requirement under `## ADDED Requirements` in a delta spec lacks the configured normative keywords
- **THEN** the validation report returns an error for the missing normative keyword

#### Scenario: Delta requirement with custom keywords passes
- **WHEN** the project config has `keywords: { normative: ["DEBE"] }` and a delta ADDED requirement contains "DEBE"
- **THEN** the validation report does not flag a missing normative keyword error

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

### Requirement: Configurable validator construction
The system SHALL support creating a validator with custom normative keywords and custom conditional keywords. The existing `NewValidator(strict bool)` constructor SHALL continue to work with the default keywords. When custom conditional keywords are provided, the validator SHALL use them in guide messages. When conditional keywords are nil, the validator SHALL use defaults (WHEN, THEN, AND).

#### Scenario: Validator with default keywords
- **WHEN** `NewValidator(strict)` is called
- **THEN** the validator checks for SHALL or MUST in requirement text and uses WHEN/THEN/AND in guide messages

#### Scenario: Validator with custom keywords
- **WHEN** `NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"}, &ConditionalsConfig{When: "CUANDO", Then: "ENTONCES", And: "Y"})` is called
- **THEN** the validator checks for DEBE or DEBERA in requirement text and uses CUANDO/ENTONCES/Y in guide messages

#### Scenario: Validator with nil keywords falls back to defaults
- **WHEN** `NewValidatorWithKeywords(false, nil, nil)` is called with nil conditionals
- **THEN** the validator checks for SHALL or MUST and uses WHEN/THEN/AND in guide messages

### Requirement: Parametrized error messages
The system SHALL generate validation error messages that include the project's configured normative keywords. The message for a missing normative keyword SHALL list the expected keywords (e.g., "Requirement must contain DEBE or DEBERA keyword" instead of the hardcoded "SHALL or MUST").

#### Scenario: Error message with default keywords
- **WHEN** a requirement fails normative keyword validation with default keywords
- **THEN** the error message reads "Requirement must contain SHALL or MUST keyword"

#### Scenario: Error message with custom keywords
- **WHEN** a requirement fails normative keyword validation with keywords ["DEBE", "DEBERA"]
- **THEN** the error message reads "Requirement must contain DEBE or DEBERA keyword"

