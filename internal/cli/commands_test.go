package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// cmdMu serializes command execution since rootCmd is a package-level singleton
// and we redirect os.Stdout/os.Stderr.
var cmdMu sync.Mutex

// resetCommandFlags resets all flag values to their defaults across all commands.
// This is necessary because rootCmd is a package-level singleton and Cobra does not
// reset flag values between Execute() calls.
func resetCommandFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, child := range cmd.Commands() {
		resetCommandFlags(child)
	}
}

// executeCommand runs a CLI command capturing stdout and stderr.
// Commands in this package write directly to os.Stdout, so we redirect at the fd level.
// It also redirects color.Output/color.Error since fatih/color caches those at init time.
func executeCommand(t *testing.T, projectRoot string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmdMu.Lock()
	defer cmdMu.Unlock()

	// Reset all flags to defaults before each execution
	resetCommandFlags(rootCmd)

	// Save and restore working directory
	origDir, dirErr := os.Getwd()
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	if chErr := os.Chdir(projectRoot); chErr != nil {
		t.Fatal(chErr)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Isolate global config
	tmpConfig := t.TempDir()
	tmpData := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpConfig)
	t.Setenv("XDG_DATA_HOME", tmpData)

	// Disable color
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	// Capture stdout
	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Redirect color.Output (fatih/color caches os.Stdout at init time)
	origColorOutput := color.Output
	color.Output = wOut
	defer func() { color.Output = origColorOutput }()

	// Capture stderr
	origStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	origColorError := color.Error
	color.Error = wErr
	defer func() { color.Error = origColorError }()

	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	wOut.Close()
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String(), execErr
}

func setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	openspecDir := filepath.Join(root, "openspec")
	for _, dir := range []string{
		openspecDir,
		filepath.Join(openspecDir, "specs"),
		filepath.Join(openspecDir, "changes"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	configContent := "schema: spec-driven\n"
	if err := os.WriteFile(filepath.Join(openspecDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeChange(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "openspec", "changes", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "proposal.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeSpec(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Join(root, "openspec", "specs", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

var sampleProposal = `# add-auth

## Why

We need authentication for security.

## What Changes

- **auth-service:** Add authentication service

## Impact

Backend API changes.
`

var sampleSpec = `# user-auth

## Purpose

Authentication system for users.

## Requirements

### Login

Users SHALL be able to login with email and password.

#### Scenario: Successful login

- **WHEN** valid credentials provided
- **THEN** user receives a session token
`

// --- list command ---

func TestList_NoOpenspecDir(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "list")
	if err == nil {
		t.Error("expected error when no openspec dir exists")
	}
	if err != nil && !strings.Contains(err.Error(), "openspec") {
		t.Errorf("expected error mentioning 'openspec', got: %v", err)
	}
}

func TestList_EmptyChanges(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No changes found") {
		t.Errorf("expected 'No changes found' in output, got: %q", stdout)
	}
}

func TestList_WithChanges(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "add-auth") {
		t.Errorf("expected 'add-auth' in output, got: %q", stdout)
	}
}

func TestList_JSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &items); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %q", jsonErr, stdout)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item in JSON, got %d", len(items))
	}
}

func TestList_Specs(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "list", "--specs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "user-auth") {
		t.Errorf("expected 'user-auth' in output, got: %q", stdout)
	}
}

func TestList_SortName(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "beta-feature", sampleProposal)
	writeChange(t, root, "alpha-feature", sampleProposal)

	stdout, _, err := executeCommand(t, root, "list", "--sort", "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	alphaIdx := strings.Index(stdout, "alpha-feature")
	betaIdx := strings.Index(stdout, "beta-feature")
	if alphaIdx < 0 || betaIdx < 0 {
		t.Fatalf("expected both features in output, got: %q", stdout)
	}
	if alphaIdx > betaIdx {
		t.Error("expected alpha-feature before beta-feature with --sort name")
	}
}

// --- show command ---

func TestShow_Change(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "show", "add-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "## Why") {
		t.Errorf("expected proposal markdown in output, got: %q", stdout)
	}
}

func TestShow_ChangeJSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "show", "add-auth", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if result["name"] != "add-auth" {
		t.Errorf("expected name 'add-auth', got %v", result["name"])
	}
}

func TestShow_Spec(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "show", "user-auth", "--type", "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "## Purpose") {
		t.Errorf("expected spec markdown in output, got: %q", stdout)
	}
}

func TestShow_NotFound(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "show", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestShow_AutoSelect(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "only-change", sampleProposal)

	stdout, _, err := executeCommand(t, root, "show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "## Why") {
		t.Errorf("expected auto-selected change content, got: %q", stdout)
	}
}

func TestShow_SpecJSON(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "show", "user-auth", "--type", "spec", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if result["name"] != "user-auth" {
		t.Errorf("expected name 'user-auth', got %v", result["name"])
	}
}

// --- validate command ---

func TestValidate_NoArgsNoFlags(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "validate")
	if err == nil {
		t.Error("expected error when no args or flags")
	}
	if err != nil && !strings.Contains(err.Error(), "specify") {
		t.Errorf("expected error about specifying args, got: %v", err)
	}
}

func TestValidate_SingleItemJSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "validate", "add-auth", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &results); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestValidate_NotFound(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "validate", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestValidate_AllJSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "validate", "--all", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &results); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestValidate_ChangesOnlyJSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "validate", "--changes", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &results); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	for _, r := range results {
		if r["type"] != "change" {
			t.Errorf("expected type 'change', got %v", r["type"])
		}
	}
}

func TestValidate_SpecsOnlyJSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, _, err := executeCommand(t, root, "validate", "--specs", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &results); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	for _, r := range results {
		if r["type"] != "spec" {
			t.Errorf("expected type 'spec', got %v", r["type"])
		}
	}
}

// --- schemas command ---

func TestSchemas_Output(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "schemas")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "spec-driven") {
		t.Errorf("expected 'spec-driven' in output, got: %q", stdout)
	}
}

func TestSchemas_JSON(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "schemas", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var schemas []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &schemas); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if len(schemas) == 0 {
		t.Error("expected at least one schema in JSON output")
	}
}

// --- completion command ---

func TestCompletion_Shells(t *testing.T) {
	root := setupProject(t)
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			stdout, _, err := executeCommand(t, root, "completion", shell)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", shell, err)
			}
			if len(stdout) == 0 {
				t.Errorf("expected non-empty completion output for %s", shell)
			}
		})
	}
}

// --- init command ---

func TestInit_CreatesStructure(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "init", "--tools", "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify directory structure
	for _, dir := range []string{"openspec", "openspec/specs", "openspec/changes"} {
		path := filepath.Join(root, dir)
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Errorf("expected directory %s to exist: %v", dir, statErr)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}

	// Verify config.yaml
	configPath := filepath.Join(root, "openspec", "config.yaml")
	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config.yaml: %v", readErr)
	}
	if !strings.Contains(string(data), "schema:") {
		t.Error("config.yaml does not contain 'schema:'")
	}
}

func TestInit_ToolsNone(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "init", "--tools", "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No tool-specific dirs should be created
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), ".") && e.Name() != ".." {
			t.Errorf("unexpected hidden directory %q created with --tools none", e.Name())
		}
	}
}

func TestInit_DoesNotOverwriteConfig(t *testing.T) {
	root := setupProject(t)
	configPath := filepath.Join(root, "openspec", "config.yaml")

	customContent := "schema: my-custom-schema\n"
	if err := os.WriteFile(configPath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeCommand(t, root, "init", "--tools", "none")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != customContent {
		t.Errorf("config.yaml was overwritten: got %q, want %q", string(data), customContent)
	}
}

func TestInit_ToolsClaude(t *testing.T) {
	root := t.TempDir()
	_, _, err := executeCommand(t, root, "init", "--tools", "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that .claude directory was created
	claudeDir := filepath.Join(root, ".claude")
	if _, statErr := os.Stat(claudeDir); os.IsNotExist(statErr) {
		t.Error("expected .claude directory to be created with --tools claude")
	}
}

// --- new change command ---

func TestNewChange_Valid(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "new", "change", "my-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "my-feature") {
		t.Errorf("expected 'my-feature' in output, got: %q", stdout)
	}

	// Verify directory was created
	changeDir := filepath.Join(root, "openspec", "changes", "my-feature")
	if _, statErr := os.Stat(changeDir); os.IsNotExist(statErr) {
		t.Error("expected change directory to be created")
	}

	// Verify proposal.md exists
	proposalPath := filepath.Join(changeDir, "proposal.md")
	if _, statErr := os.Stat(proposalPath); os.IsNotExist(statErr) {
		t.Error("expected proposal.md to be created")
	}
}

func TestNewChange_AlreadyExists(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "existing", sampleProposal)

	_, _, err := executeCommand(t, root, "new", "change", "existing")
	if err == nil {
		t.Error("expected error when change already exists")
	}
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

func TestNewChange_InvalidName(t *testing.T) {
	root := setupProject(t)
	_, _, err := executeCommand(t, root, "new", "change", "Invalid Name")
	if err == nil {
		t.Error("expected error for invalid change name")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected 'invalid' in error, got: %v", err)
	}
}

// --- status command ---

func TestStatus_NoChanges(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No active changes") {
		t.Errorf("expected 'No active changes' in output, got: %q", stdout)
	}
}

func TestStatus_WithChange(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "status", "--change", "add-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "add-auth") {
		t.Errorf("expected 'add-auth' in output, got: %q", stdout)
	}
}

func TestStatus_JSON(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, _, err := executeCommand(t, root, "status", "--change", "add-auth", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if result["change"] != "add-auth" {
		t.Errorf("expected change 'add-auth', got %v", result["change"])
	}
}

// --- templates command ---

func TestTemplates_Output(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "proposal") {
		t.Errorf("expected 'proposal' in output, got: %q", stdout)
	}
}

func TestTemplates_JSON(t *testing.T) {
	root := setupProject(t)
	stdout, _, err := executeCommand(t, root, "templates", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &result); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}
	if _, ok := result["proposal"]; !ok {
		t.Error("expected 'proposal' key in JSON output")
	}
}

// --- deprecated commands ---

func TestDeprecated_ChangeList(t *testing.T) {
	root := setupProject(t)
	writeChange(t, root, "add-auth", sampleProposal)

	stdout, stderr, err := executeCommand(t, root, "change", "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce same JSON output as "list"
	var items []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &items); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}

	// Stderr should mention "deprecated"
	if !strings.Contains(strings.ToLower(stderr), "deprecated") {
		t.Errorf("expected 'deprecated' warning in stderr, got: %q", stderr)
	}
}

func TestDeprecated_SpecList(t *testing.T) {
	root := setupProject(t)
	writeSpec(t, root, "user-auth", sampleSpec)

	stdout, stderr, err := executeCommand(t, root, "spec", "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []interface{}
	if jsonErr := json.Unmarshal([]byte(stdout), &items); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %q", jsonErr, stdout)
	}

	if !strings.Contains(strings.ToLower(stderr), "deprecated") {
		t.Errorf("expected 'deprecated' warning in stderr, got: %q", stderr)
	}
}
