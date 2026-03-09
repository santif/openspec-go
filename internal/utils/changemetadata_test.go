package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadChangeMetadata_Valid(t *testing.T) {
	dir := t.TempDir()
	content := "schema: spec-driven\ncreated: \"2025-01-15T10:30:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".openspec.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	meta, err := ReadChangeMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Schema != "spec-driven" {
		t.Errorf("expected schema 'spec-driven', got %q", meta.Schema)
	}
	if meta.Created != "2025-01-15T10:30:00Z" {
		t.Errorf("expected created '2025-01-15T10:30:00Z', got %q", meta.Created)
	}
}

func TestReadChangeMetadata_MissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := ReadChangeMetadata(dir)
	if err == nil {
		t.Error("expected error for missing .openspec.yaml file")
	}
}

func TestReadChangeMetadata_OnlySchema(t *testing.T) {
	dir := t.TempDir()
	content := "schema: custom-schema\n"
	if err := os.WriteFile(filepath.Join(dir, ".openspec.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	meta, err := ReadChangeMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Schema != "custom-schema" {
		t.Errorf("expected schema 'custom-schema', got %q", meta.Schema)
	}
	if meta.Created != "" {
		t.Errorf("expected empty created field, got %q", meta.Created)
	}
}

func TestWriteChangeMetadata(t *testing.T) {
	dir := t.TempDir()

	original := &ChangeMetadata{
		Schema:  "spec-driven",
		Created: "2025-06-01T08:00:00Z",
	}
	if err := WriteChangeMetadata(dir, original); err != nil {
		t.Fatalf("unexpected error writing metadata: %v", err)
	}

	meta, err := ReadChangeMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error reading metadata: %v", err)
	}
	if meta.Schema != original.Schema {
		t.Errorf("schema mismatch: expected %q, got %q", original.Schema, meta.Schema)
	}
	if meta.Created != original.Created {
		t.Errorf("created mismatch: expected %q, got %q", original.Created, meta.Created)
	}
}

func TestWriteChangeMetadata_CreatesFile(t *testing.T) {
	dir := t.TempDir()

	meta := &ChangeMetadata{
		Schema:  "my-schema",
		Created: "2025-03-10T12:00:00Z",
	}
	if err := WriteChangeMetadata(dir, meta); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metaPath := filepath.Join(dir, ".openspec.yaml")
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("expected .openspec.yaml to exist at %s, got error: %v", metaPath, err)
	}
}
