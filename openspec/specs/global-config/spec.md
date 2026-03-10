## Purpose
XDG-compliant global configuration system that stores user preferences in `~/.config/openspec/config.json`, managing workflow profiles, delivery mode, feature flags, and providing CLI commands for inspection and modification.

## Requirements

### Requirement: XDG-compliant storage
The system SHALL store global configuration at the XDG-compliant path `~/.config/openspec/config.json` on Linux and macOS. On Windows, the system SHALL use the `%APPDATA%` directory. The system MUST create the directory if it does not exist.

#### Scenario: Config file location on Linux/macOS
- **WHEN** global config is loaded on a Linux or macOS system
- **THEN** the system reads from `~/.config/openspec/config.json`

#### Scenario: Config directory creation
- **WHEN** global config is saved and `~/.config/openspec/` does not exist
- **THEN** the system creates the directory before writing the file

### Requirement: Configuration fields
The system SHALL support the following configuration fields: profile (string: "core" or "custom"), delivery (string: "skills", "commands", or "both"), features (map of string to boolean for feature flags), and workflows (list of active workflow names).

#### Scenario: Load config with all fields
- **WHEN** config.json contains profile="custom", delivery="both", features={"telemetry": false}, workflows=["propose","apply"]
- **THEN** `Load()` returns a GlobalConfig with all fields populated correctly

#### Scenario: Load config with missing optional fields
- **WHEN** config.json contains only profile="core"
- **THEN** `Load()` returns a GlobalConfig with defaults for delivery, features, and workflows

### Requirement: Workflow profiles
The system SHALL define two profiles: `core` with 4 workflows (propose, explore, apply, archive) and `custom` with 11 workflows (adding new, continue, ff, sync, bulk-archive, verify, onboard). The active profile SHALL determine which workflows are available for command and skill generation.

#### Scenario: Core profile workflows
- **WHEN** the global config has profile="core"
- **THEN** only propose, explore, apply, and archive workflows are active

#### Scenario: Custom profile workflows
- **WHEN** the global config has profile="custom"
- **THEN** all 11 workflows are active

### Requirement: Config CLI commands
The system SHALL provide the following `openspec config` subcommands: `path` (display config file path), `list` (display all settings), `get <key>` (display single setting), `set <key> <value>` (update a setting), `unset <key>` (remove a setting), `reset` (restore defaults), and `edit` (open in $EDITOR).

#### Scenario: Get a config value
- **WHEN** `openspec config get profile` is run
- **THEN** the current profile value is displayed

#### Scenario: Set a config value
- **WHEN** `openspec config set delivery skills` is run
- **THEN** the delivery field is updated to "skills" and saved to disk

#### Scenario: Reset config to defaults
- **WHEN** `openspec config reset` is run
- **THEN** the config file is overwritten with default values

#### Scenario: Edit config in editor
- **WHEN** `openspec config edit` is run
- **THEN** the system opens the config file in the user's $EDITOR

### Requirement: Default values
The system SHALL provide sensible defaults when the config file does not exist: profile defaults to "core", delivery defaults to "both", features defaults to an empty map, and workflows defaults to the core workflow list.

#### Scenario: First-time load with no config file
- **WHEN** `Load()` is called and `~/.config/openspec/config.json` does not exist
- **THEN** the returned config has profile="core", delivery="both", empty features, and core workflows

### Requirement: Config persistence
The system SHALL atomically read and write the JSON config file. Writes MUST serialize the full config state. Reads MUST parse the JSON and return a structured config object.

#### Scenario: Save and reload config
- **WHEN** a config is saved with profile="custom" and then loaded again
- **THEN** the loaded config has profile="custom"
