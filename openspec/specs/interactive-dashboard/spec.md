## ADDED Requirements

### Requirement: Overview display
The system SHALL display specs, active changes, and archived changes in a single terminal view when `openspec view` is run. Each section SHALL be clearly labeled and separated.

#### Scenario: Display full project overview
- **WHEN** `openspec view` is run with 3 specs, 2 active changes, and 1 archived change
- **THEN** the terminal displays all three sections with their respective items

#### Scenario: Display with no active changes
- **WHEN** `openspec view` is run with specs but no active changes
- **THEN** the specs section is displayed and the active changes section shows an empty state message

### Requirement: Progress visualization
The system SHALL display an overall progress bar per change showing completed vs total artifacts. The progress bar MUST use Unicode characters to visualize the ratio of completed to total artifacts. Per-artifact done/ready/blocked display is provided by the `status` command, not `view`.

#### Scenario: Show progress for a change
- **WHEN** `openspec view` renders a change with 2 of 4 artifacts completed
- **THEN** the progress bar shows 50% completion with 2 filled and 2 empty segments

### Requirement: Task progress display
The system SHALL display completed and total task counts from tasks.md for each change that has a tasks file. The format MUST show `X/Y tasks` where X is completed and Y is total.

#### Scenario: Display task progress
- **WHEN** a change has tasks.md with 3 completed and 5 total tasks
- **THEN** the dashboard shows "3/5 tasks" for that change

#### Scenario: No tasks file
- **WHEN** a change has no tasks.md file
- **THEN** the dashboard omits task progress for that change

### Requirement: Color-coded status indicators
The system SHALL use terminal colors to indicate change-level completion: green "OK" when all artifacts are complete, and yellow ".." for changes that are in-progress. The indicators reflect overall change completion, not per-artifact status.

#### Scenario: Fully complete change shown in green
- **WHEN** all artifacts for a change are completed
- **THEN** the change's status indicator shows green "OK"

#### Scenario: In-progress change shown in yellow
- **WHEN** a change has some artifacts completed but not all
- **THEN** the change's status indicator shows yellow ".."

### Requirement: Empty state handling
The system SHALL display "(none)" text when no specs, changes, or archived changes exist in a section.

#### Scenario: Empty project
- **WHEN** `openspec view` is run on a project with no specs and no changes
- **THEN** the dashboard shows "(none)" under each section
