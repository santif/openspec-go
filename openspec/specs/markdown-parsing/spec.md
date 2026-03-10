## ADDED Requirements

### Requirement: Generic section parsing
The system SHALL parse any markdown document into a hierarchical tree of sections. Each section SHALL have a Level (determined by the number of `#` characters), Title (the heading text), Content (body text between this heading and the next), and Children (nested subsections). The parser SHALL support heading levels 1 through 6.

#### Scenario: Parse a document with nested sections
- **WHEN** a markdown document contains `## Parent` followed by content and `### Child` followed by content
- **THEN** the parser returns a section tree where Parent is at level 2 with Child as a nested child at level 3, each with their respective content

#### Scenario: Parse a flat document with no headings
- **WHEN** a markdown document contains only body text with no headings
- **THEN** the parser returns an empty slice because content without headings is discarded

#### Scenario: Find a section by title
- **WHEN** `FindSection()` is called with a title that exists in the parsed tree
- **THEN** the system returns the matching section and its children

### Requirement: Spec parsing
The system SHALL parse spec markdown files extracting a `Purpose` section and a `Requirements` section. Requirements SHALL be identified as level-3 headings (`###`) under the Requirements section. Each requirement MAY contain scenarios as level-4 headings (`####`).

#### Scenario: Parse a valid spec with requirements and scenarios
- **WHEN** a spec file contains `## Purpose`, `## Requirements`, `### Requirement: Auth`, and `#### Scenario: Login`
- **THEN** `ParseSpec()` returns a Spec with the purpose text, one requirement named "Auth" with one scenario containing the raw text

#### Scenario: Parse a spec missing the Purpose section
- **WHEN** a spec file contains `## Requirements` but no `## Purpose` section
- **THEN** `ParseSpec()` returns an error indicating the Purpose section is missing

### Requirement: Change parsing
The system SHALL parse change markdown files extracting a `Why` section and a `What Changes` section. The What Changes section SHALL contain delta bullets in the format `- **SpecName:** description` or `- **SpecName**: description` (both colon formats accepted), each parsed into a Delta with the spec name and description.

#### Scenario: Parse a change with delta bullets
- **WHEN** a change file contains `## Why` with motivation text and `## What Changes` with `- **UserAuth:** Add MFA support`
- **THEN** `ParseChange()` returns a Change with the why text and one delta referencing spec "UserAuth" with description "Add MFA support"

#### Scenario: Parse a change with no delta bullets
- **WHEN** a change file contains `## Why` and `## What Changes` with plain text but no bold-prefixed bullets
- **THEN** `ParseChange()` returns a Change with zero deltas

### Requirement: Delta spec parsing
The system SHALL parse delta spec files containing operation headers (`## ADDED Requirements`, `## MODIFIED Requirements`, `## REMOVED Requirements`, `## RENAMED Requirements`). Under each operation header, requirement blocks SHALL be extracted with their names, content, and scenarios.

#### Scenario: Parse a delta spec with ADDED and REMOVED operations
- **WHEN** a delta spec file contains `## ADDED Requirements` with `### Requirement: NewFeature` and `## REMOVED Requirements` with `### Requirement: OldFeature`
- **THEN** `ParseDeltaSpec()` returns two operation groups: ADDED containing "NewFeature" and REMOVED containing "OldFeature"

#### Scenario: Normalize requirement names for matching
- **WHEN** comparing requirement names "  User Auth " and "user auth"
- **THEN** `NormalizeRequirementName()` produces the same normalized string for both, enabling case-insensitive and whitespace-insensitive matching

### Requirement: Requirements section extraction
The system SHALL structurally extract the Requirements section from a spec file, splitting it into Before (content above Requirements), HeaderLine (the `## Requirements` heading), Preamble (text between the heading and first requirement), BodyBlocks (ordered list of requirement blocks), and After (content below Requirements).

#### Scenario: Extract requirements section from a full spec
- **WHEN** `ExtractRequirementsSection()` is called on a spec with Purpose, Requirements (containing two requirements), and an appendix section
- **THEN** the result contains the Purpose in Before, the two requirements as BodyBlocks, and the appendix in After

#### Scenario: Extract from a spec with preamble text
- **WHEN** the Requirements section has introductory text before the first `### Requirement:` heading
- **THEN** that text is captured in the Preamble field, separate from BodyBlocks

### Requirement: Filesystem-aware change parsing
The system SHALL scan a change directory's `specs/` subdirectory for delta spec files at `specs/<name>/spec.md`. Each found file SHALL be parsed and attached to the corresponding delta in the change's What Changes section, building a complete change with all delta details.

#### Scenario: Parse a change with delta spec files
- **WHEN** `ParseChangeWithDeltas()` is called on a change directory containing `proposal.md` and `specs/user-auth/spec.md`
- **THEN** the returned Change includes the parsed delta spec for "user-auth" attached to the matching delta

#### Scenario: Parse a change with no specs directory
- **WHEN** `ParseChangeWithDeltas()` is called on a change directory that has no `specs/` subdirectory
- **THEN** the returned Change has deltas from proposal.md only, with no delta spec details

### Requirement: Requirement name normalization
The system SHALL normalize requirement names by trimming whitespace and comparing case-insensitively. This normalization MUST be used consistently when matching requirements across delta operations (MODIFIED, REMOVED, RENAMED).

#### Scenario: Match requirements with different casing
- **WHEN** a MODIFIED operation targets "User Authentication" and the existing spec has "user authentication"
- **THEN** the system matches them as the same requirement

#### Scenario: Match requirements with extra whitespace
- **WHEN** a REMOVED operation targets "  Data Export  " and the existing spec has "Data Export"
- **THEN** the system matches them as the same requirement

### Requirement: Delta operation auto-detection
The system SHALL automatically detect the operation type when parsing delta bullets from the What Changes section. The detection MUST be based on description keywords: "add", "create", or "new" map to ADDED; "remove" or "delete" map to REMOVED; "rename" map to RENAMED; all other descriptions default to MODIFIED. Keyword matching SHALL be case-insensitive.

#### Scenario: Detect ADDED operation from description
- **WHEN** a delta bullet contains `- **NewFeature:** Add authentication support`
- **THEN** the parsed delta has operation type ADDED based on the "Add" keyword

#### Scenario: Default to MODIFIED for unrecognized descriptions
- **WHEN** a delta bullet contains `- **Auth:** Improve error handling`
- **THEN** the parsed delta has operation type MODIFIED as the default

### Requirement: Case-insensitive section finding
The system SHALL match section titles case-insensitively when using `FindSection()`. The comparison MUST use lowercased versions of both the search title and actual section titles.

#### Scenario: Find section regardless of case
- **WHEN** `FindSection()` is called with title "purpose" on a document containing `## Purpose`
- **THEN** the system returns the matching section
