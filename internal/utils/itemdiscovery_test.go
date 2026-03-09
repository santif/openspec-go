package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetActiveChangeIDs_Empty(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatal(err)
	}

	ids := GetActiveChangeIDs(dir)
	if len(ids) != 0 {
		t.Errorf("expected empty slice, got %v", ids)
	}
}

func TestGetActiveChangeIDs_FindsChanges(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes")

	// Create two change subdirectories
	for _, name := range []string{"beta-feature", "alpha-feature"} {
		if err := os.MkdirAll(filepath.Join(changesDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	ids := GetActiveChangeIDs(dir)
	if len(ids) != 2 {
		t.Fatalf("expected 2 changes, got %d: %v", len(ids), ids)
	}
	if ids[0] != "alpha-feature" {
		t.Errorf("expected first element 'alpha-feature', got %q", ids[0])
	}
	if ids[1] != "beta-feature" {
		t.Errorf("expected second element 'beta-feature', got %q", ids[1])
	}
}

func TestGetActiveChangeIDs_IgnoresHidden(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes")

	// Create a visible and a hidden subdirectory
	for _, name := range []string{"visible-change", ".hidden"} {
		if err := os.MkdirAll(filepath.Join(changesDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	ids := GetActiveChangeIDs(dir)
	if len(ids) != 1 {
		t.Fatalf("expected 1 change, got %d: %v", len(ids), ids)
	}
	if ids[0] != "visible-change" {
		t.Errorf("expected 'visible-change', got %q", ids[0])
	}
}

func TestGetActiveChangeIDs_NoDirReturnsNil(t *testing.T) {
	dir := t.TempDir()
	// Do not create openspec/changes/ directory

	ids := GetActiveChangeIDs(dir)
	if ids != nil {
		t.Errorf("expected nil, got %v", ids)
	}
}

func TestGetSpecIDs(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "openspec", "specs")

	for _, name := range []string{"user-auth", "billing", "notifications"} {
		if err := os.MkdirAll(filepath.Join(specsDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	ids := GetSpecIDs(dir)
	if len(ids) != 3 {
		t.Fatalf("expected 3 specs, got %d: %v", len(ids), ids)
	}
	// Results should be sorted
	expected := []string{"billing", "notifications", "user-auth"}
	for i, want := range expected {
		if ids[i] != want {
			t.Errorf("ids[%d] = %q, want %q", i, ids[i], want)
		}
	}
}

func TestGetArchivedChangeIDs(t *testing.T) {
	dir := t.TempDir()
	archiveDir := filepath.Join(dir, "openspec", "archive")

	for _, name := range []string{"done-feature", "archived-change"} {
		if err := os.MkdirAll(filepath.Join(archiveDir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}

	ids := GetArchivedChangeIDs(dir)
	if len(ids) != 2 {
		t.Fatalf("expected 2 archived changes, got %d: %v", len(ids), ids)
	}
	// Results should be sorted
	if ids[0] != "archived-change" {
		t.Errorf("expected first element 'archived-change', got %q", ids[0])
	}
	if ids[1] != "done-feature" {
		t.Errorf("expected second element 'done-feature', got %q", ids[1])
	}
}
