package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	if FileExists(file) {
		t.Error("expected file to not exist")
	}

	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if !FileExists(file) {
		t.Error("expected file to exist")
	}
}

func TestDirectoryExists(t *testing.T) {
	dir := t.TempDir()

	if !DirectoryExists(dir) {
		t.Error("expected directory to exist")
	}

	if DirectoryExists(filepath.Join(dir, "nonexistent")) {
		t.Error("expected directory to not exist")
	}

	// File should not pass directory check
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if DirectoryExists(file) {
		t.Error("expected file to not pass as directory")
	}
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")

	if err := EnsureDir(nested); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !DirectoryExists(nested) {
		t.Error("expected nested directory to be created")
	}

	// Should not error on existing dir
	if err := EnsureDir(nested); err != nil {
		t.Fatalf("unexpected error on existing dir: %v", err)
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sub", "test.txt")

	if err := WriteFile(file, "hello world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	content, err := ReadFile(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "test content" {
		t.Errorf("expected 'test content', got %q", content)
	}

	// Non-existent file
	_, err = ReadFile(filepath.Join(dir, "nonexistent.txt"))
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestUpdateFileWithMarkers_NewFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	err := UpdateFileWithMarkers(file, "injected content", "<!-- START -->", "<!-- END -->")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if !strings.Contains(content, "injected content") {
		t.Error("expected injected content")
	}
	if !strings.Contains(content, "<!-- START -->") {
		t.Error("expected start marker")
	}
	if !strings.Contains(content, "<!-- END -->") {
		t.Error("expected end marker")
	}
}

func TestUpdateFileWithMarkers_ReplaceContent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	// Write initial content
	initial := "<!-- START -->\nold content\n<!-- END -->"
	if err := WriteFile(file, initial); err != nil {
		t.Fatal(err)
	}

	err := UpdateFileWithMarkers(file, "new content", "<!-- START -->", "<!-- END -->")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if strings.Contains(content, "old content") {
		t.Error("expected old content to be replaced")
	}
	if !strings.Contains(content, "new content") {
		t.Error("expected new content")
	}
}

func TestUpdateFileWithMarkers_PreserveAround(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	initial := "before content\n<!-- START -->\nold\n<!-- END -->\nafter content"
	if err := WriteFile(file, initial); err != nil {
		t.Fatal(err)
	}

	err := UpdateFileWithMarkers(file, "new", "<!-- START -->", "<!-- END -->")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if !strings.Contains(content, "before content") {
		t.Error("expected content before markers to be preserved")
	}
	if !strings.Contains(content, "after content") {
		t.Error("expected content after markers to be preserved")
	}
}

func TestUpdateFileWithMarkers_UnpairedMarker(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	// Only start marker
	if err := WriteFile(file, "<!-- START -->\nsome content"); err != nil {
		t.Fatal(err)
	}

	err := UpdateFileWithMarkers(file, "new", "<!-- START -->", "<!-- END -->")
	if err == nil {
		t.Error("expected error for unpaired marker")
	}
}

func TestUpdateFileWithMarkers_Idempotent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	content := "same content"
	if err := UpdateFileWithMarkers(file, content, "<!-- S -->", "<!-- E -->"); err != nil {
		t.Fatal(err)
	}
	first, _ := ReadFile(file)

	if err := UpdateFileWithMarkers(file, content, "<!-- S -->", "<!-- E -->"); err != nil {
		t.Fatal(err)
	}
	second, _ := ReadFile(file)

	if first != second {
		t.Errorf("expected idempotent result\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestUpdateFileWithMarkers_EndBeforeStart(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")

	// End marker before start marker — should error
	if err := WriteFile(file, "<!-- END -->\nsome content\n<!-- START -->"); err != nil {
		t.Fatal(err)
	}

	err := UpdateFileWithMarkers(file, "new", "<!-- START -->", "<!-- END -->")
	if err == nil {
		t.Error("expected error when end marker appears before start marker")
	}
}

func TestDetectShell_WithSHELLEnv(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	result := DetectShell()
	if result != "zsh" {
		t.Errorf("expected 'zsh', got %q", result)
	}
}

func TestDetectShell_EmptySHELL(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fallback shell differs on Windows")
	}
	t.Setenv("SHELL", "")
	result := DetectShell()
	// On non-Windows, should fall back to "bash"
	if result != "bash" {
		t.Errorf("expected 'bash' fallback, got %q", result)
	}
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "deeply", "nested", "dir", "file.txt")

	if err := WriteFile(file, "content"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("expected 'content', got %q", string(data))
	}
}

func TestReadChangeMetadata_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".openspec.yaml"), []byte("invalid: yaml: [broken"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadChangeMetadata(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
