## ADDED Requirements

### Requirement: Legacy cleanup
The system SHALL detect and clean legacy OpenSpec patterns from the project. This includes removing old slash-command directories and OpenSpec marker blocks from AI tool configuration files. The cleanup MUST only remove patterns matching known legacy markers defined in `config.OpenSpecMarkers`.

#### Scenario: Detect legacy slash-command directories
- **WHEN** legacy cleanup is run and the project contains an old-format skill directory
- **THEN** the system detects and removes the legacy directory

#### Scenario: Remove marker blocks from config files
- **WHEN** legacy cleanup is run and an AI tool config file contains an OpenSpec marker block
- **THEN** the system removes the marker block from the config file, preserving other content

#### Scenario: No legacy patterns found
- **WHEN** legacy cleanup is run on a project with no legacy patterns
- **THEN** the system reports no changes needed

### Requirement: Profile migration
The system SHALL perform one-time profile migration when the config format changes between versions. `MigrateIfNeeded()` SHALL detect whether migration is required by checking for the presence of skill files on disk and global config state. The migration MUST set up the global config profile and delivery mode based on detected skill file presence.

#### Scenario: Migrate from legacy config
- **WHEN** `MigrateIfNeeded()` runs and detects legacy workflow configuration
- **THEN** the system migrates to the new format, setting profile and delivery mode in global config

#### Scenario: No migration needed
- **WHEN** `MigrateIfNeeded()` runs and the config is already up to date
- **THEN** the system skips migration and returns immediately

### Requirement: Profile sync drift detection
The system SHALL detect when project skill/command files on disk drift from the desired configuration in global config. Drift occurs when workflows are added or removed from the profile, delivery mode changes, or skill files are manually modified. The detection MUST signal that `openspec update` is needed to resynchronize.

#### Scenario: Detect missing skill files
- **WHEN** the global config has profile="custom" (11 workflows) but only 4 skill files exist on disk
- **THEN** drift detection reports that skill files are out of sync

#### Scenario: Detect extra skill files
- **WHEN** the global config has profile="core" (4 workflows) but 11 skill files exist on disk
- **THEN** drift detection reports that extra skill files exist beyond the configured profile

#### Scenario: No drift detected
- **WHEN** skill files on disk match the configured profile and delivery mode exactly
- **THEN** drift detection reports no drift

### Requirement: JSON conversion
The system SHALL convert spec and change markdown files to structured JSON for programmatic consumption. `ConvertSpecToJSON()` SHALL produce a JSON object with purpose, requirements (each with name, text, scenarios), and metadata. `ConvertChangeToJSON()` SHALL produce a JSON object with why, what_changes, and deltas.

#### Scenario: Convert a spec to JSON
- **WHEN** `ConvertSpecToJSON()` is called on a spec with 2 requirements and 3 scenarios
- **THEN** the returned JSON has purpose text and a requirements array with 2 entries totaling 3 scenarios

#### Scenario: Convert a change to JSON
- **WHEN** `ConvertChangeToJSON()` is called on a change with 2 deltas
- **THEN** the returned JSON has why text, what_changes text, and a deltas array with 2 entries
