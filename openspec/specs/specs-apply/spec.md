## Purpose
Spec application engine that discovers delta spec files from a change, merges ADDED/MODIFIED/REMOVED/RENAMED operations into main specs with atomic validation-then-write, supports dry-run mode, and reports operation counts.

## Requirements

### Requirement: Spec update discovery
The system SHALL scan a change's `specs/` directory for delta spec files at `specs/<name>/spec.md`. Each found file SHALL be matched to an existing spec in `openspec/specs/<name>/spec.md` or flagged as a new spec creation.

#### Scenario: Discover delta specs for existing specs
- **WHEN** `FindSpecUpdates()` scans a change with `specs/user-auth/spec.md` and `openspec/specs/user-auth/spec.md` exists
- **THEN** the result includes user-auth as an update to an existing spec

#### Scenario: Discover delta specs for new specs
- **WHEN** `FindSpecUpdates()` scans a change with `specs/new-feature/spec.md` and no matching spec exists in `openspec/specs/`
- **THEN** the result includes new-feature as a new spec creation

### Requirement: Skeleton creation for new specs
The system SHALL generate new spec files from delta specs that contain only ADDED requirements. The skeleton MUST include a `## Purpose` section (empty) and a `## Requirements` header only — requirement content is NOT pre-populated in the skeleton. Requirements are added during `BuildUpdatedSpec()` processing.

#### Scenario: Create spec from ADDED-only delta
- **WHEN** `BuildSpecSkeleton()` is called with a delta containing only ADDED requirements
- **THEN** the output is a complete spec file with Purpose and Requirements sections containing the added requirements

#### Scenario: Reject non-ADDED operations for new specs
- **WHEN** a delta for a new spec contains MODIFIED operations and `BuildUpdatedSpec()` processes it
- **THEN** `BuildUpdatedSpec()` returns a validation error because MODIFIED operations require an existing spec with matching requirements

### Requirement: Delta merging with correct operation order
The system SHALL apply delta operations to existing spec content in the fixed order: RENAMED first, then REMOVED, then MODIFIED, then ADDED. This ordering MUST be enforced regardless of the order operations appear in the delta spec file.

#### Scenario: Apply operations in correct order
- **WHEN** a delta spec contains ADDED, REMOVED, MODIFIED, and RENAMED operations
- **THEN** `BuildUpdatedSpec()` processes them in fixed order: RENAMED first, then REMOVED, then MODIFIED, then ADDED

#### Scenario: Rename then modify
- **WHEN** a delta renames "Old Name" to "New Name" and modifies "New Name"
- **THEN** the system first renames the requirement, then applies the modification to the renamed requirement

### Requirement: Delta validation rules
The system SHALL enforce validation rules before applying deltas. The system MUST reject: duplicate requirement names within ADDED, MODIFIED targeting a non-existent requirement, REMOVED targeting a non-existent requirement, RENAMED targeting a non-existent requirement, RENAMED with a TO name that already exists, a new spec containing only REMOVED operations, conflicting operations on the same requirement, and duplicate requirements in the resulting spec.

#### Scenario: Reject MODIFIED targeting non-existent requirement
- **WHEN** a delta has a MODIFIED operation for "Feature X" but the existing spec has no requirement named "Feature X"
- **THEN** the system returns a validation error indicating the target requirement was not found

#### Scenario: Reject RENAMED to an existing name
- **WHEN** a delta renames "Old Feature" to "Existing Feature" and "Existing Feature" already exists in the spec
- **THEN** the system returns a validation error indicating the target name already exists

#### Scenario: Reject new spec with only REMOVED operations
- **WHEN** a delta for a non-existent spec contains only REMOVED requirements
- **THEN** the system returns a validation error because there is no spec to remove from

#### Scenario: Reject duplicate ADDED requirements
- **WHEN** a delta has two ADDED requirements both named "Feature A"
- **THEN** the system returns a validation error indicating duplicate requirement names

### Requirement: Atomic application
The system SHALL validate all spec updates before writing any files. If any single spec update fails validation, no files SHALL be written. The system MUST use a validate-all-then-write-all strategy.

#### Scenario: Atomic rollback on validation failure
- **WHEN** three specs are being updated and the second fails validation
- **THEN** none of the three spec files are written to disk

#### Scenario: All valid specs written
- **WHEN** all spec updates pass validation
- **THEN** all updated spec files are written to disk

### Requirement: Operation counting
The system SHALL track and report the count of each operation type (added, modified, removed, renamed) across all applied delta specs. This summary MUST be returned after a successful application.

#### Scenario: Report operation counts
- **WHEN** `ApplySpecs()` successfully applies deltas with 3 ADDED, 1 MODIFIED, and 1 REMOVED operations
- **THEN** the returned summary includes added=3, modified=1, removed=1, renamed=0

### Requirement: Dry-run mode
The system SHALL support a dry-run mode that computes and validates all spec updates without writing any files to the filesystem. The dry run MUST report what would be changed.

#### Scenario: Dry run reports changes without writing
- **WHEN** `ApplySpecs()` is called in dry-run mode with valid deltas
- **THEN** the system reports what changes would be applied but no files on disk are modified
