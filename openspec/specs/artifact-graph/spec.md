## Purpose
Dependency graph system that constructs a DAG from schema artifact definitions, providing topological ordering, blocked-artifact detection, completion state tracking, and next-ready identification to orchestrate artifact generation order.

## Requirements

### Requirement: Graph construction from schema
The system SHALL construct a directed acyclic graph (DAG) from a schema's artifact definitions, where each artifact is a node and each entry in its `requires` list creates an edge from the dependency to the dependent.

#### Scenario: Build graph from default schema
- **WHEN** `NewGraphFromSchema()` is called with a schema defining proposal→specs→design→tasks dependencies
- **THEN** the graph contains four nodes with edges reflecting the dependency chain

#### Scenario: Build graph from schema with no dependencies
- **WHEN** a schema has artifacts with empty requires lists
- **THEN** the graph contains all artifacts as independent nodes with no edges

### Requirement: Topological ordering
The system SHALL produce a deterministic topological ordering of artifacts using Kahn's algorithm. Artifacts with no unmet dependencies SHALL appear first.

#### Scenario: Get build order for linear dependency chain
- **WHEN** `GetBuildOrder()` is called on a graph with proposal→specs→design→tasks
- **THEN** the returned order is [proposal, specs, design, tasks] (or any valid topological sort)

#### Scenario: Get build order with parallel artifacts
- **WHEN** specs and design both depend only on proposal, and tasks depends on both
- **THEN** proposal appears first, specs and design appear in any order after proposal, and tasks appears last

### Requirement: Blocked artifact detection
The system SHALL identify artifacts that cannot proceed because one or more of their dependencies are not yet completed. The result MUST include which specific dependencies are blocking each artifact.

#### Scenario: Detect blocked artifact
- **WHEN** `GetBlocked()` is called and specs is not yet completed
- **THEN** design and tasks are reported as blocked, with specs listed as the blocking dependency for design

#### Scenario: No blocked artifacts when all completed
- **WHEN** all artifacts are completed
- **THEN** `GetBlocked()` returns an empty list

### Requirement: Completion detection via filesystem
The system SHALL detect artifact completion by scanning the change directory for files matching the artifact's `generates` glob pattern using doublestar matching. An artifact is complete only if at least one matching file exists and is non-empty.

#### Scenario: Detect completed artifact
- **WHEN** `DetectCompleted()` runs and a change directory contains a non-empty `proposal.md` matching the proposal artifact's generates pattern
- **THEN** the proposal artifact is marked as completed

#### Scenario: Empty file does not count as complete
- **WHEN** a change directory contains `proposal.md` but the file is empty (0 bytes)
- **THEN** the proposal artifact is NOT marked as completed

#### Scenario: Glob matching for specs artifact
- **WHEN** the specs artifact has generates pattern `specs/**/*.md` and the change directory contains `specs/auth/spec.md`
- **THEN** the specs artifact is marked as completed

### Requirement: Task progress tracking
The system SHALL parse `tasks.md` files for checkbox-format tasks (`- [x]` or `- [X]` for completed, `- [ ]` for pending) and return total and completed counts. The line MUST be lowercased before regex matching, so both `[x]` and `[X]` are accepted as completed.

#### Scenario: Count tasks in a tasks.md file
- **WHEN** `CountTasks()` is called on a tasks.md with 3 `- [x]` and 2 `- [ ]` items
- **THEN** the result shows total=5, completed=3

#### Scenario: No tasks file exists
- **WHEN** `CountTasks()` is called but tasks.md does not exist in the change directory
- **THEN** the result shows total=0, completed=0

### Requirement: Status display
The system SHALL provide an `openspec status` command that displays per-artifact status (done, ready, or blocked) for a specific change. The display SHALL include task progress when available and support JSON output.

#### Scenario: Show status for a change with mixed states
- **WHEN** `openspec status --change my-change` is run and proposal is done, specs is ready, design is blocked
- **THEN** the output shows proposal as done, specs as ready, and design as blocked by specs

#### Scenario: JSON status output
- **WHEN** `openspec status --change my-change --json` is run
- **THEN** the output is a JSON object with artifact statuses, task progress, and blocked dependencies

### Requirement: Next-ready artifact detection
The system SHALL provide `GetNextArtifacts()` to return artifacts whose dependencies are all completed, identifying what can be worked on next. An artifact is next-ready when it is not yet completed and all artifacts in its `requires` list are completed.

#### Scenario: Identify next-ready artifacts
- **WHEN** `GetNextArtifacts()` is called and proposal is completed but specs, design, tasks are not
- **THEN** specs is returned as the next-ready artifact (its only dependency, proposal, is completed)

#### Scenario: No next-ready artifacts when blocked
- **WHEN** no incomplete artifact has all its dependencies completed
- **THEN** `GetNextArtifacts()` returns an empty list

### Requirement: Graph completeness check
The system SHALL provide `IsComplete()` to check whether all artifacts in the graph are completed. The method MUST return true only when every artifact has been marked as completed.

#### Scenario: Graph is complete
- **WHEN** `IsComplete()` is called and all artifacts are completed
- **THEN** the method returns true

#### Scenario: Graph is not complete
- **WHEN** `IsComplete()` is called and one artifact is not completed
- **THEN** the method returns false
