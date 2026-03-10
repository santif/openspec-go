## MODIFIED Requirements

### Requirement: Spec validation
The system SHALL validate spec files by checking for required sections (Purpose, Requirements), verifying that requirements are at heading level 3 (`###`), verifying that scenarios are at heading level 4 (`####`), and checking for the presence of configured normative keywords in requirement text. When no custom keywords are configured, the system SHALL default to requiring SHALL or MUST.

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

### Requirement: Delta requirement content validation
The system SHALL validate requirements within ADDED and MODIFIED delta sections using the same rules as main spec requirements. Each requirement MUST contain the configured normative keywords and MUST have at least one scenario. When no custom keywords are configured, the system SHALL default to requiring SHALL or MUST.

#### Scenario: Delta requirement missing normative keyword
- **WHEN** a requirement under `## ADDED Requirements` in a delta spec lacks the configured normative keywords
- **THEN** the validation report returns an error for the missing normative keyword

#### Scenario: Delta requirement with custom keywords passes
- **WHEN** the project config has `keywords: { normative: ["DEBE"] }` and a delta ADDED requirement contains "DEBE"
- **THEN** the validation report does not flag a missing normative keyword error

## ADDED Requirements

### Requirement: Configurable validator construction
The system SHALL support creating a validator with custom normative keywords via a new constructor `NewValidatorWithKeywords(strict bool, keywords []string)`. The existing `NewValidator(strict bool)` constructor SHALL continue to work with the default keywords (SHALL, MUST). When custom keywords are provided, the validator SHALL build a word-boundary regex dynamically from the keyword list.

#### Scenario: Validator with default keywords
- **WHEN** `NewValidator(strict)` is called
- **THEN** the validator checks for SHALL or MUST in requirement text

#### Scenario: Validator with custom keywords
- **WHEN** `NewValidatorWithKeywords(false, []string{"DEBE", "DEBERA"})` is called
- **THEN** the validator checks for DEBE or DEBERA in requirement text

#### Scenario: Validator with nil keywords falls back to defaults
- **WHEN** `NewValidatorWithKeywords(false, nil)` is called
- **THEN** the validator checks for SHALL or MUST (default behavior)

### Requirement: Parametrized error messages
The system SHALL generate validation error messages that include the project's configured normative keywords. The message for a missing normative keyword SHALL list the expected keywords (e.g., "Requirement must contain DEBE or DEBERA keyword" instead of the hardcoded "SHALL or MUST").

#### Scenario: Error message with default keywords
- **WHEN** a requirement fails normative keyword validation with default keywords
- **THEN** the error message reads "Requirement must contain SHALL or MUST keyword"

#### Scenario: Error message with custom keywords
- **WHEN** a requirement fails normative keyword validation with keywords ["DEBE", "DEBERA"]
- **THEN** the error message reads "Requirement must contain DEBE or DEBERA keyword"
