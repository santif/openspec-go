## MODIFIED Requirements

### Requirement: Config structure
The system SHALL store per-project configuration in a YAML file at `openspec/config.yaml`. The file SHALL support the following fields: schema (string), profile (string), workflows (list of strings), context (string), rules (map of artifact ID to arrays of rule strings), and keywords (object with normative list of strings and conditionals object with named keys: when, then, and).

#### Scenario: Parse a complete config file
- **WHEN** `ReadProjectConfig()` is called on a config.yaml with all fields populated including `keywords.normative` and `keywords.conditionals`
- **THEN** the returned ProjectConfig has schema, profile, workflows, context, rules, and keywords correctly set including conditionals with when, then, and and values

#### Scenario: Parse a minimal config file
- **WHEN** `ReadProjectConfig()` is called on a config.yaml with only `schema: spec-driven`
- **THEN** the returned ProjectConfig has schema="spec-driven" with defaults for all other fields, keywords is nil

#### Scenario: Parse config with custom normative keywords
- **WHEN** config.yaml contains `keywords: { normative: ["DEBE", "DEBERA"] }`
- **THEN** the returned ProjectConfig has Keywords.Normative set to ["DEBE", "DEBERA"] and Keywords.Conditionals is nil

#### Scenario: Parse config with empty normative keywords list
- **WHEN** config.yaml contains `keywords: { normative: [] }`
- **THEN** the returned ProjectConfig has Keywords with an empty Normative slice and Conditionals is nil

#### Scenario: Parse config with custom conditional keywords
- **WHEN** config.yaml contains `keywords: { conditionals: { when: "CUANDO", then: "ENTONCES", and: "Y" } }`
- **THEN** the returned ProjectConfig has Keywords.Conditionals with When="CUANDO", Then="ENTONCES", And="Y"

#### Scenario: Parse config with partial conditional keywords
- **WHEN** config.yaml contains `keywords: { conditionals: { when: "CUANDO" } }` with then and and omitted
- **THEN** the returned ProjectConfig has Keywords.Conditionals with When="CUANDO", Then="" and And=""

### Requirement: Keywords validation
The system SHALL validate the `keywords.normative` field when present. Each keyword MUST be a non-empty string. The system SHALL warn when a keyword contains regex metacharacters to prevent regex injection. The system SHALL warn when the keywords list is present but empty (validator will fall back to defaults). The system SHALL validate `keywords.conditionals` when present: the system SHALL warn when a named key (when, then, or and) is empty, and SHALL warn when a named key contains regex metacharacters.

#### Scenario: Valid custom keywords
- **WHEN** config.yaml has `keywords: { normative: ["DEBE", "DEBERA"] }`
- **THEN** validation passes with no issues

#### Scenario: Keyword with regex metacharacter
- **WHEN** config.yaml has `keywords: { normative: ["MUST(not)"] }`
- **THEN** validation returns a warning about unsafe characters in keyword

#### Scenario: Empty keywords list
- **WHEN** config.yaml has `keywords: { normative: [] }`
- **THEN** validation returns a warning that normative keywords list is empty (validator will fall back to defaults)

#### Scenario: Valid conditional keywords
- **WHEN** config.yaml has `keywords: { conditionals: { when: "CUANDO", then: "ENTONCES", and: "Y" } }`
- **THEN** validation passes with no issues

#### Scenario: Conditional keywords with empty value
- **WHEN** config.yaml has `keywords: { conditionals: { when: "", then: "ENTONCES", and: "Y" } }`
- **THEN** validation returns a warning that conditionals.when is empty

#### Scenario: Conditional keywords with regex metacharacter
- **WHEN** config.yaml has `keywords: { conditionals: { when: "WHEN(ever)", then: "THEN", and: "AND" } }`
- **THEN** validation returns a warning about unsafe characters in conditional keyword

## ADDED Requirements

### Requirement: Conditionals config defaults
The system SHALL provide a method to resolve effective conditional keywords. When `Keywords.Conditionals` is nil, the effective values SHALL be the defaults: When="WHEN", Then="THEN", And="AND". When `Keywords.Conditionals` is set, partially configured fields SHALL be merged with defaults (empty strings are replaced with default values).

#### Scenario: Resolve defaults when conditionals not configured
- **WHEN** `Keywords` is nil and `ResolveConditionals()` is called
- **THEN** the returned values are When="WHEN", Then="THEN", And="AND"

#### Scenario: Resolve configured conditionals
- **WHEN** `Keywords.Conditionals` is set with When="CUANDO", Then="ENTONCES", And="Y"
- **THEN** `ResolveConditionals()` returns the configured values

#### Scenario: Resolve partially configured conditionals
- **WHEN** `Keywords.Conditionals` is set with only When="CUANDO" (Then and And are empty)
- **THEN** `ResolveConditionals()` returns When="CUANDO", Then="THEN", And="AND" (empty fields merged with defaults)
