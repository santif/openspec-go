## MODIFIED Requirements

### Requirement: Config structure
The system SHALL store per-project configuration in a YAML file at `openspec/config.yaml`. The file SHALL support the following fields: schema (string), profile (string), workflows (list of strings), context (string), rules (map of artifact ID to arrays of rule strings), and keywords (object with normative list of strings).

#### Scenario: Parse a complete config file
- **WHEN** `Load()` is called on a config.yaml with all fields populated including `keywords.normative`
- **THEN** the returned ProjectConfig has schema, profile, workflows, context, rules, and keywords correctly set

#### Scenario: Parse a minimal config file
- **WHEN** `Load()` is called on a config.yaml with only `schema: spec-driven`
- **THEN** the returned ProjectConfig has schema="spec-driven" with defaults for all other fields and keywords is nil

#### Scenario: Parse config with custom normative keywords
- **WHEN** config.yaml contains `keywords: { normative: ["DEBE", "DEBERA"] }`
- **THEN** the returned ProjectConfig has Keywords.Normative set to ["DEBE", "DEBERA"]

#### Scenario: Parse config with empty normative keywords list
- **WHEN** config.yaml contains `keywords: { normative: [] }`
- **THEN** the returned ProjectConfig has Keywords with an empty Normative slice

## ADDED Requirements

### Requirement: Keywords validation
The system SHALL validate the `keywords.normative` field when present. Each keyword MUST be a non-empty string. The system SHALL reject keywords that contain regex metacharacters to prevent regex injection. The system SHALL warn when the keywords list is present but empty.

#### Scenario: Valid custom keywords
- **WHEN** config.yaml has `keywords: { normative: ["DEBE", "DEBERA"] }`
- **THEN** validation passes with no issues

#### Scenario: Keyword with regex metacharacter
- **WHEN** config.yaml has `keywords: { normative: ["MUST(not)"] }`
- **THEN** validation returns a warning about unsafe characters in keyword

#### Scenario: Empty keywords list
- **WHEN** config.yaml has `keywords: { normative: [] }`
- **THEN** validation returns a warning that normative keywords list is empty (validator will fall back to defaults)
