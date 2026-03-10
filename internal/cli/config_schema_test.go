package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- config command ---

func TestConfig_Path(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "openspec") {
		t.Errorf("expected path containing 'openspec', got: %q", stdout)
	}
}

func TestConfig_List(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("expected valid JSON output, got parse error: %v\nraw: %q", jsonErr, stdout)
	}
}

func TestConfig_Get_Profile(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "get", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "core" {
		t.Errorf("expected default profile 'core', got: %q", strings.TrimSpace(stdout))
	}
}

func TestConfig_Get_Delivery(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "get", "delivery")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "both" {
		t.Errorf("expected default delivery 'both', got: %q", strings.TrimSpace(stdout))
	}
}

func TestConfig_Get_Workflows(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "config", "get", "workflows")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Get_UnknownKey(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "config", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown config key")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' in error, got: %v", err)
	}
}

func TestConfig_Set_Profile(t *testing.T) {
	root := setupProject(t)

	stdout, _, err := executeCommand(t, root, "config", "set", "profile", "custom")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Set profile = custom") {
		t.Errorf("expected 'Set profile = custom' in output, got: %q", stdout)
	}
}

func TestConfig_Set_Delivery(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "set", "delivery", "skills")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Set delivery = skills") {
		t.Errorf("expected 'Set delivery = skills' in output, got: %q", stdout)
	}
}

func TestConfig_Set_UnknownKey(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "config", "set", "nonexistent", "val")
	if err == nil {
		t.Fatal("expected error for unknown config key")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' in error, got: %v", err)
	}
}

func TestConfig_Unset_Profile(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "unset", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Unset profile") {
		t.Errorf("expected 'Unset profile' in output, got: %q", stdout)
	}
}

func TestConfig_Unset_Workflows(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "unset", "workflows")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Unset workflows") {
		t.Errorf("expected 'Unset workflows' in output, got: %q", stdout)
	}
}

func TestConfig_Reset(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "config", "reset")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Configuration reset to defaults") {
		t.Errorf("expected 'Configuration reset to defaults' in output, got: %q", stdout)
	}
}

// --- schema command ---

func TestSchema_Which(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "schema", "which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "spec-driven" {
		t.Errorf("expected 'spec-driven', got: %q", strings.TrimSpace(stdout))
	}
}

func TestSchema_Validate(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "schema", "validate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(stdout), "valid") {
		t.Errorf("expected output containing 'valid', got: %q", stdout)
	}
}

func TestSchema_Validate_WithArg(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "schema", "validate", "spec-driven")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(stdout), "valid") {
		t.Errorf("expected output containing 'valid', got: %q", stdout)
	}
}

func TestSchema_Fork(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "schema", "fork")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	forkedDir := filepath.Join(root, "openspec", "schemas", "spec-driven")
	info, statErr := os.Stat(forkedDir)
	if statErr != nil {
		t.Fatalf("expected forked schema directory to exist: %v", statErr)
	}
	if !info.IsDir() {
		t.Error("expected forked schema path to be a directory")
	}
}

func TestSchema_Fork_AlreadyExists(t *testing.T) {
	root := setupProject(t)

	// Pre-create the schema directory
	forkedDir := filepath.Join(root, "openspec", "schemas", "spec-driven")
	if err := os.MkdirAll(forkedDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeCommand(t, root, "schema", "fork")
	if err == nil {
		t.Fatal("expected error when schema already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}
