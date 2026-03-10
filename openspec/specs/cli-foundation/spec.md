## ADDED Requirements

### Requirement: Root command and entry point
The system SHALL provide a root Cobra command with `Execute()` as the entry point called from `main.go`. The root command SHALL display usage and help information when run without subcommands.

#### Scenario: Run without subcommands
- **WHEN** `openspec` is run with no arguments
- **THEN** the system displays usage information listing available subcommands

#### Scenario: Display help
- **WHEN** `openspec --help` is run
- **THEN** the system displays detailed help including all commands and global flags

### Requirement: Version injection
The system SHALL accept a version string injected via Go ldflags at build time (`-X github.com/santif/openspec-go/internal/cli.version=VERSION`). The `--version` flag SHALL display the version in the format "openspec version X.Y.Z".

#### Scenario: Display injected version
- **WHEN** the binary is built with `-X ...version=1.2.3` and `openspec --version` is run
- **THEN** the output displays "openspec version 1.2.3"

#### Scenario: Default version when not injected
- **WHEN** the binary is built without ldflags and `openspec --version` is run
- **THEN** the output includes a default version string (e.g., "dev")

### Requirement: Deprecated command aliases
The system SHALL provide backward-compatible aliases for `openspec change [show|list|validate]` and `openspec spec [show|list|validate]`. Each alias MUST display a deprecation warning directing users to the new command.

#### Scenario: Use deprecated change show alias
- **WHEN** `openspec change show my-change` is run
- **THEN** the system executes `openspec show my-change` and displays a deprecation warning

#### Scenario: Use deprecated spec validate alias
- **WHEN** `openspec spec validate my-spec` is run
- **THEN** the system executes `openspec validate my-spec` as a spec and displays a deprecation warning

### Requirement: Feedback submission
The system SHALL provide an `openspec feedback <message>` command that creates a GitHub issue via the `gh` CLI. A `--body` flag SHALL allow providing a detailed description.

#### Scenario: Submit feedback
- **WHEN** `openspec feedback "Feature request: add diff view"` is run with `gh` available
- **THEN** a GitHub issue is created in the configured repository with the message as the title

#### Scenario: Submit feedback with body
- **WHEN** `openspec feedback "Bug report" --body "Steps to reproduce..."` is run
- **THEN** a GitHub issue is created with the message as title and body as description

### Requirement: Filesystem utilities
The system SHALL provide utility functions for common filesystem operations: `EnsureDir` (create directory if not exists), `FileExists` (check file existence), `DirectoryExists` (check directory existence), `ReadFile` (read file contents), `WriteFile` (write file contents), `UpdateFileWithMarkers` (update file content between marker comments), and `RemoveMarkerBlock` (remove a marker block from a file). Move and remove operations use `os.Rename()` and `os.Remove()` directly where needed.

#### Scenario: EnsureDir creates missing directory
- **WHEN** `EnsureDir("/path/to/new/dir")` is called and the directory does not exist
- **THEN** the directory and all parent directories are created

#### Scenario: FileExists detects existing file
- **WHEN** `FileExists("/path/to/file")` is called on an existing file
- **THEN** the function returns true

#### Scenario: FileExists returns false for missing file
- **WHEN** `FileExists("/path/to/missing")` is called on a non-existent path
- **THEN** the function returns false

### Requirement: Cross-platform support
The system SHALL support building for linux, darwin, and windows on both amd64 and arm64 architectures. GoReleaser SHALL be configured to produce binaries for all target combinations.

#### Scenario: Build for all platforms
- **WHEN** `goreleaser release` is run
- **THEN** binaries are produced for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, and windows/arm64
