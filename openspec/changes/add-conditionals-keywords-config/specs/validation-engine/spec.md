## MODIFIED Requirements

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
