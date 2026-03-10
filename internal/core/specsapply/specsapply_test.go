package specsapply

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santif/openspec-go/internal/utils"
)

// --- Helpers ---

func setupChangeDir(t *testing.T, specName, deltaContent string) string {
	t.Helper()
	dir := t.TempDir()
	specDir := filepath.Join(dir, "specs", specName)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func setupMainSpec(t *testing.T, dir, specName, content string) {
	t.Helper()
	specDir := filepath.Join(dir, specName)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func makeUpdate(changeDir, mainSpecsDir, specName string) SpecUpdate {
	return SpecUpdate{
		Source: filepath.Join(changeDir, "specs", specName, "spec.md"),
		Target: filepath.Join(mainSpecsDir, specName, "spec.md"),
		Exists: utils.FileExists(filepath.Join(mainSpecsDir, specName, "spec.md")),
	}
}

// --- FindSpecUpdates ---

func TestFindSpecUpdates_NoSpecsDir(t *testing.T) {
	dir := t.TempDir()
	updates, err := FindSpecUpdates(dir, filepath.Join(dir, "main"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 0 {
		t.Errorf("expected 0 updates, got %d", len(updates))
	}
}

func TestFindSpecUpdates_FindsDeltaSpecs(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", "## ADDED Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	mainDir := t.TempDir()

	updates, err := FindSpecUpdates(changeDir, mainDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].Exists {
		t.Error("expected Exists to be false for new spec")
	}
}

func TestFindSpecUpdates_DetectsExistence(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", "## ADDED Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth system.\n\n## Requirements\n")

	updates, err := FindSpecUpdates(changeDir, mainDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if !updates[0].Exists {
		t.Error("expected Exists to be true for existing spec")
	}
}

// --- BuildSpecSkeleton ---

func TestBuildSpecSkeleton(t *testing.T) {
	skeleton := BuildSpecSkeleton("auth", "add-auth")
	if !strings.Contains(skeleton, "# auth Specification") {
		t.Error("expected skeleton to contain spec title")
	}
	if !strings.Contains(skeleton, "## Purpose") {
		t.Error("expected skeleton to contain Purpose section")
	}
	if !strings.Contains(skeleton, "## Requirements") {
		t.Error("expected skeleton to contain Requirements section")
	}
	if !strings.Contains(skeleton, "add-auth") {
		t.Error("expected skeleton to reference change name")
	}
}

// --- BuildUpdatedSpec ---

func TestBuildUpdatedSpec_AddedToNewSpec(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## ADDED Requirements

### Requirement: Login
The system SHALL allow users to login.

#### Scenario: Basic login
- **WHEN** user enters valid credentials
- **THEN** system grants access
`)
	mainDir := t.TempDir()
	update := makeUpdate(changeDir, mainDir, "auth")

	rebuilt, counts, err := BuildUpdatedSpec(update, "add-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Added != 1 {
		t.Errorf("expected 1 added, got %d", counts.Added)
	}
	if !strings.Contains(rebuilt, "### Requirement: Login") {
		t.Error("expected rebuilt spec to contain added requirement")
	}
	if !strings.Contains(rebuilt, "## Purpose") {
		t.Error("expected rebuilt spec to contain Purpose from skeleton")
	}
}

func TestBuildUpdatedSpec_RemovedIgnoredForNewSpec(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## ADDED Requirements

### Requirement: Login
The system SHALL allow users to login.

#### Scenario: Basic login
- **WHEN** user enters valid credentials
- **THEN** system grants access

## REMOVED Requirements

### Requirement: Old Feature
`)
	mainDir := t.TempDir()
	update := makeUpdate(changeDir, mainDir, "auth")

	rebuilt, counts, err := BuildUpdatedSpec(update, "add-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Added != 1 {
		t.Errorf("expected 1 added, got %d", counts.Added)
	}
	if counts.Removed != 1 {
		t.Errorf("expected 1 removed in counts, got %d", counts.Removed)
	}
	if !strings.Contains(rebuilt, "### Requirement: Login") {
		t.Error("expected rebuilt spec to contain the added requirement")
	}
}

func TestBuildUpdatedSpec_ModifiedErrorForNewSpec(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: Login
The system SHALL do something modified.

#### Scenario: Test
- **WHEN** test
`)
	mainDir := t.TempDir()
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "add-auth")
	if err == nil {
		t.Fatal("expected error for MODIFIED on new spec")
	}
	if !strings.Contains(err.Error(), "only ADDED requirements are allowed for new specs") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_RenamedErrorForNewSpec(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: Old`"+`
- TO: `+"`### Requirement: New`"+`
`)
	mainDir := t.TempDir()
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "add-auth")
	if err == nil {
		t.Fatal("expected error for RENAMED on new spec")
	}
	if !strings.Contains(err.Error(), "only ADDED requirements are allowed for new specs") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_OperationOrder(t *testing.T) {
	existingSpec := `# Auth Specification

## Purpose
Authentication and authorization system for the application.

## Requirements

### Requirement: A
The system SHALL do A.

#### Scenario: A test
- **WHEN** A
- **THEN** A works

### Requirement: B
The system SHALL do B.

#### Scenario: B test
- **WHEN** B
- **THEN** B works

### Requirement: C-old
The system SHALL do C with old name.

#### Scenario: C test
- **WHEN** C
- **THEN** C works
`

	deltaContent := `## RENAMED Requirements

- FROM: ` + "`### Requirement: A`" + `
- TO: ` + "`### Requirement: A-renamed`" + `

## REMOVED Requirements

### Requirement: B

## MODIFIED Requirements

### Requirement: A-renamed
The system SHALL do A-renamed with updates.

#### Scenario: A-renamed test
- **WHEN** A-renamed
- **THEN** A-renamed works

## ADDED Requirements

### Requirement: D
The system SHALL do D.

#### Scenario: D test
- **WHEN** D
- **THEN** D works
`

	changeDir := setupChangeDir(t, "auth", deltaContent)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", existingSpec)
	update := makeUpdate(changeDir, mainDir, "auth")

	rebuilt, counts, err := BuildUpdatedSpec(update, "test-change")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Renamed != 1 {
		t.Errorf("expected 1 renamed, got %d", counts.Renamed)
	}
	if counts.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", counts.Removed)
	}
	if counts.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", counts.Modified)
	}
	if counts.Added != 1 {
		t.Errorf("expected 1 added, got %d", counts.Added)
	}

	// B should be removed
	if strings.Contains(rebuilt, "### Requirement: B") {
		t.Error("expected B to be removed")
	}
	// A should be renamed
	if strings.Contains(rebuilt, "### Requirement: A\n") {
		t.Error("expected A to be renamed")
	}
	if !strings.Contains(rebuilt, "### Requirement: A-renamed") {
		t.Error("expected A-renamed to exist")
	}
	// D should be added
	if !strings.Contains(rebuilt, "### Requirement: D") {
		t.Error("expected D to be added")
	}
	// C-old should remain unchanged
	if !strings.Contains(rebuilt, "### Requirement: C-old") {
		t.Error("expected C-old to remain")
	}
}

func TestBuildUpdatedSpec_DuplicateInAdded(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## ADDED Requirements

### Requirement: Login
SHALL login.

#### Scenario: T
- **WHEN** test

### Requirement: Login
SHALL login again.

#### Scenario: T2
- **WHEN** test
`)
	mainDir := t.TempDir()
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for duplicate in ADDED")
	}
	if !strings.Contains(err.Error(), "duplicate requirement in ADDED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_DuplicateInModified(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: Login
SHALL login v1.

#### Scenario: T
- **WHEN** test

### Requirement: Login
SHALL login v2.

#### Scenario: T2
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for duplicate in MODIFIED")
	}
	if !strings.Contains(err.Error(), "duplicate requirement in MODIFIED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_DuplicateInRemoved(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## REMOVED Requirements

### Requirement: Login
### Requirement: Login
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for duplicate in REMOVED")
	}
	if !strings.Contains(err.Error(), "duplicate requirement in REMOVED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_DuplicateInRenamedFrom(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: A`"+`
- TO: `+"`### Requirement: B`"+`

- FROM: `+"`### Requirement: A`"+`
- TO: `+"`### Requirement: C`"+`
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: A\nSHALL A.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for duplicate FROM in RENAMED")
	}
	if !strings.Contains(err.Error(), "duplicate FROM in RENAMED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_DuplicateInRenamedTo(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: A`"+`
- TO: `+"`### Requirement: C`"+`

- FROM: `+"`### Requirement: B`"+`
- TO: `+"`### Requirement: C`"+`
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: A\nSHALL A.\n\n#### Scenario: TA\n- **WHEN** A\n\n### Requirement: B\nSHALL B.\n\n#### Scenario: TB\n- **WHEN** B\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for duplicate TO in RENAMED")
	}
	if !strings.Contains(err.Error(), "duplicate TO in RENAMED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_CrossSectionConflict_ModifiedAndRemoved(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: Login
SHALL login modified.

#### Scenario: T
- **WHEN** test

## REMOVED Requirements

### Requirement: Login
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for cross-section conflict")
	}
	if !strings.Contains(err.Error(), "MODIFIED") || !strings.Contains(err.Error(), "REMOVED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_CrossSectionConflict_ModifiedAndAdded(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: Login
SHALL login modified.

#### Scenario: T
- **WHEN** test

## ADDED Requirements

### Requirement: Login
SHALL login added.

#### Scenario: T
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for cross-section conflict")
	}
	if !strings.Contains(err.Error(), "MODIFIED") || !strings.Contains(err.Error(), "ADDED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_CrossSectionConflict_AddedAndRemoved(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## ADDED Requirements

### Requirement: Login
SHALL login added.

#### Scenario: T
- **WHEN** test

## REMOVED Requirements

### Requirement: Login
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for cross-section conflict")
	}
	if !strings.Contains(err.Error(), "ADDED") || !strings.Contains(err.Error(), "REMOVED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_ModifiedMustReferenceNewName(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: Old`"+`
- TO: `+"`### Requirement: New`"+`

## MODIFIED Requirements

### Requirement: Old
SHALL do modified old.

#### Scenario: T
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Old\nSHALL old.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error: MODIFIED must reference new name")
	}
	if !strings.Contains(err.Error(), "MODIFIED must reference the NEW header") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_RenamedToCollidesWithAdded(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: Old`"+`
- TO: `+"`### Requirement: Collision`"+`

## ADDED Requirements

### Requirement: Collision
SHALL collide.

#### Scenario: T
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Old\nSHALL old.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for RENAMED TO colliding with ADDED")
	}
	if !strings.Contains(err.Error(), "RENAMED TO header collides with ADDED") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_NoDeltaOperations(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## Some Random Section

Just some text, no delta operations.
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for no delta operations")
	}
	if !strings.Contains(err.Error(), "no operations") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_ModifiedNotFound(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: NonExistent
SHALL modify non-existent.

#### Scenario: T
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for MODIFIED not found")
	}
	if !strings.Contains(err.Error(), "MODIFIED failed") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_RemovedNotFound(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## REMOVED Requirements

### Requirement: NonExistent
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for REMOVED not found")
	}
	if !strings.Contains(err.Error(), "REMOVED failed") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_AddedAlreadyExists(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## ADDED Requirements

### Requirement: Login
SHALL login again.

#### Scenario: T
- **WHEN** test
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for ADDED already exists")
	}
	if !strings.Contains(err.Error(), "ADDED failed") && !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_RenamedSourceNotFound(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: NonExistent`"+`
- TO: `+"`### Requirement: New`"+`
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for RENAMED source not found")
	}
	if !strings.Contains(err.Error(), "RENAMED failed") && !strings.Contains(err.Error(), "source not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_RenamedTargetAlreadyExists(t *testing.T) {
	changeDir := setupChangeDir(t, "auth", `## RENAMED Requirements

- FROM: `+"`### Requirement: Login`"+`
- TO: `+"`### Requirement: Logout`"+`
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n\n### Requirement: Logout\nSHALL logout.\n\n#### Scenario: T\n- **WHEN** test\n")
	update := makeUpdate(changeDir, mainDir, "auth")

	_, _, err := BuildUpdatedSpec(update, "test")
	if err == nil {
		t.Fatal("expected error for RENAMED target already exists")
	}
	if !strings.Contains(err.Error(), "RENAMED failed") && !strings.Contains(err.Error(), "target already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildUpdatedSpec_HeaderNormalization(t *testing.T) {
	// Test that trailing spaces in requirement names are normalized
	existingSpec := "## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login  \nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n"
	changeDir := setupChangeDir(t, "auth", `## MODIFIED Requirements

### Requirement: Login
SHALL login modified.

#### Scenario: T
- **WHEN** test updated
`)
	mainDir := t.TempDir()
	setupMainSpec(t, mainDir, "auth", existingSpec)
	update := makeUpdate(changeDir, mainDir, "auth")

	rebuilt, counts, err := BuildUpdatedSpec(update, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", counts.Modified)
	}
	if !strings.Contains(rebuilt, "SHALL login modified") {
		t.Error("expected modified content in rebuilt spec")
	}
}

// --- ApplySpecs ---

func TestApplySpecs_AtomicFailure(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")
	specsDir := filepath.Join(root, "openspec", "specs")

	// Set up change with two specs: one valid, one invalid
	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "valid-spec"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "valid-spec", "spec.md"), []byte(`## ADDED Requirements

### Requirement: Feature
The system SHALL provide a feature.

#### Scenario: Test
- **WHEN** user requests feature
- **THEN** system provides it
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid: MODIFIED on non-existing spec with no target
	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "invalid-spec"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "invalid-spec", "spec.md"), []byte(`## MODIFIED Requirements

### Requirement: NonExistent
SHALL modify non-existent.

#### Scenario: T
- **WHEN** test
`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ApplySpecs(root, "test-change", ApplyOptions{SkipValidation: true})
	if err == nil {
		t.Fatal("expected error for atomic failure")
	}

	// Neither spec should have been written
	if utils.FileExists(filepath.Join(specsDir, "valid-spec", "spec.md")) {
		t.Error("valid spec should not have been written due to atomic failure")
	}
}

func TestApplySpecs_AggregatedTotals(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")

	// First spec: 1 added
	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "spec-a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "spec-a", "spec.md"), []byte(`## ADDED Requirements

### Requirement: Feature A
The system SHALL provide feature A.

#### Scenario: Test A
- **WHEN** user requests A
- **THEN** system provides A
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Second spec: 1 added
	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "spec-b"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "spec-b", "spec.md"), []byte(`## ADDED Requirements

### Requirement: Feature B
The system SHALL provide feature B.

#### Scenario: Test B
- **WHEN** user requests B
- **THEN** system provides B
`), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := ApplySpecs(root, "test-change", ApplyOptions{SkipValidation: true, Silent: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Totals.Added != 2 {
		t.Errorf("expected total 2 added, got %d", output.Totals.Added)
	}
	if len(output.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(output.Capabilities))
	}
}

func TestApplySpecs_DryRun(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")

	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "auth", "spec.md"), []byte(`## ADDED Requirements

### Requirement: Login
The system SHALL allow users to login.

#### Scenario: Basic login
- **WHEN** user enters valid credentials
- **THEN** system grants access
`), 0644); err != nil {
		t.Fatal(err)
	}

	specsDir := filepath.Join(root, "openspec", "specs")

	output, err := ApplySpecs(root, "test-change", ApplyOptions{DryRun: true, SkipValidation: true, Silent: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.NoChanges {
		t.Error("expected NoChanges to be false")
	}
	// File should NOT have been written
	if utils.FileExists(filepath.Join(specsDir, "auth", "spec.md")) {
		t.Error("spec should not have been written in dry-run mode")
	}
}

// --- WriteUpdatedSpec ---

func TestWriteUpdatedSpec(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "specs", "auth", "spec.md")
	update := SpecUpdate{
		Source: "/some/source",
		Target: target,
	}

	content := "# Auth Specification\n\n## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n"
	counts := ApplyResult{Added: 1}

	err := WriteUpdatedSpec(update, content, counts, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !utils.FileExists(target) {
		t.Fatal("expected spec file to be written")
	}
	written, err := utils.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if written != content {
		t.Error("written content does not match expected")
	}
}

func TestWriteUpdatedSpec_Verbose(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "specs", "auth", "spec.md")
	update := SpecUpdate{
		Source: "/some/source",
		Target: target,
	}

	content := "# Auth Specification\n\n## Purpose\nAuth.\n\n## Requirements\n\n### Requirement: Login\nSHALL login.\n\n#### Scenario: T\n- **WHEN** test\n"
	counts := ApplyResult{Added: 1, Modified: 1, Removed: 1, Renamed: 1}

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := WriteUpdatedSpec(update, content, counts, false)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Applying changes to") {
		t.Errorf("expected output to contain 'Applying changes to', got: %s", output)
	}
	if !strings.Contains(output, "+ 1 added") {
		t.Errorf("expected output to contain '+ 1 added', got: %s", output)
	}
	if !strings.Contains(output, "~ 1 modified") {
		t.Errorf("expected output to contain '~ 1 modified', got: %s", output)
	}
	if !strings.Contains(output, "- 1 removed") {
		t.Errorf("expected output to contain '- 1 removed', got: %s", output)
	}
	if !strings.Contains(output, "1 renamed") {
		t.Errorf("expected output to contain '1 renamed', got: %s", output)
	}
}

func TestApplySpecs_ChangeNotFound(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := ApplySpecs(root, "nonexistent-change", ApplyOptions{})
	if err == nil {
		t.Fatal("expected error for non-existent change")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain 'not found', got: %v", err)
	}
}

func TestApplySpecs_NoSpecUpdates(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a proposal.md so the change dir exists but has no specs/ subdir
	if err := os.WriteFile(filepath.Join(changesDir, "proposal.md"), []byte("# Proposal\n"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := ApplySpecs(root, "test-change", ApplyOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.NoChanges {
		t.Error("expected NoChanges to be true")
	}
}

func TestApplySpecs_WithValidation(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")

	// Delta that produces invalid spec (requirement with no proper scenario)
	deltaContent := `## ADDED Requirements

### Requirement: Bad
No normative keywords here.
`
	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "bad-spec"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "bad-spec", "spec.md"), []byte(deltaContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ApplySpecs(root, "test-change", ApplyOptions{SkipValidation: false})
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(err.Error(), "validation errors") {
		t.Errorf("expected error to contain 'validation errors', got: %v", err)
	}
}

func TestApplySpecs_DryRun_Verbose(t *testing.T) {
	root := t.TempDir()
	changesDir := filepath.Join(root, "openspec", "changes", "test-change")

	if err := os.MkdirAll(filepath.Join(changesDir, "specs", "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "specs", "auth", "spec.md"), []byte(`## ADDED Requirements

### Requirement: Login
The system SHALL allow users to login.

#### Scenario: Basic login
- **WHEN** user enters valid credentials
- **THEN** system grants access
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	output, err := ApplySpecs(root, "test-change", ApplyOptions{DryRun: true, Silent: false, SkipValidation: true})

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	captured := buf.String()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.NoChanges {
		t.Error("expected NoChanges to be false")
	}
	if !strings.Contains(captured, "Would apply changes to") {
		t.Errorf("expected output to contain 'Would apply changes to', got: %s", captured)
	}

	// File should NOT have been written
	specsDir := filepath.Join(root, "openspec", "specs")
	if utils.FileExists(filepath.Join(specsDir, "auth", "spec.md")) {
		t.Error("spec should not have been written in dry-run mode")
	}
}
