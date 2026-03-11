## Purpose
Per-project configuration system that reads `openspec/config.yaml` to control schema selection, profile overrides, workflow lists, context injection, rules mapping, and keyword customization.
## Requirements
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

### Requirement: Schema selection
The system SHALL use the `schema` field to determine which schema the project uses. This value MUST match a schema name resolvable through the schema resolution chain.

#### Scenario: Config references a valid schema
- **WHEN** config.yaml has `schema: spec-driven` and the schema resolution chain can resolve "spec-driven"
- **THEN** the project uses the resolved schema for all operations

### Requirement: Context field
The system SHALL support a free-text `context` field for project-level context that gets injected into AI instructions. The system MUST enforce a 50KB size limit on the context field value.

#### Scenario: Context within size limit
- **WHEN** config.yaml has a context field with 10KB of text
- **THEN** `Load()` succeeds and the context is available for instruction enrichment

#### Scenario: Context exceeds size limit
- **WHEN** config.yaml has a context field exceeding 50KB
- **THEN** the system prints a warning to stderr and silently ignores the oversized context, continuing without error

### Requirement: Rules mapping
The system SHALL support a `rules` field mapping artifact IDs to arrays of rule strings (`map[string][]string`). When an instruction is enriched for an artifact, the corresponding rules (if any) SHALL be injected into the instruction.

#### Scenario: Rule for specs artifact
- **WHEN** config.yaml has `rules: { specs: ["All requirements must include performance criteria"] }`
- **THEN** the "specs" rules are available for injection during instruction enrichment

#### Scenario: No rule for an artifact
- **WHEN** an artifact has no corresponding entry in the rules map
- **THEN** no rule text is injected for that artifact

### Requirement: Profile override
The system SHALL allow the project-level config to override the global profile. When a project config specifies a profile, it MUST take precedence over the global config profile.

#### Scenario: Project profile overrides global
- **WHEN** global config has profile="core" and project config has profile="custom"
- **THEN** the effective profile for the project is "custom"

#### Scenario: No project profile falls through to global
- **WHEN** project config has no profile field and global config has profile="core"
- **THEN** the effective profile is "core"

### Requirement: Config loading with defaults
The system SHALL parse the YAML config file. `Load()` SHALL return `nil` when the config has only minimal content with all-empty fields. The default schema SHALL be "spec-driven". The default for context SHALL be empty string. The default for rules SHALL be an empty map.

#### Scenario: Load config with defaults applied
- **WHEN** `Load()` is called on a config.yaml with only `schema: spec-driven`
- **THEN** context defaults to empty, rules defaults to empty map, profile defaults to empty (falls through to global)

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

