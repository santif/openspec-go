package projectconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadProjectConfig_YAMLFile(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
profile: core
workflows:
  - propose
  - explore
  - apply
context: "This is the project context for testing."
rules:
  proposal:
    - "Keep proposals concise"
    - "Include rationale"
  spec:
    - "Use SHALL for requirements"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.Schema != "spec-driven" {
		t.Errorf("expected schema 'spec-driven', got %q", config.Schema)
	}
	if config.Profile != "core" {
		t.Errorf("expected profile 'core', got %q", config.Profile)
	}
	if len(config.Workflows) != 3 {
		t.Errorf("expected 3 workflows, got %d", len(config.Workflows))
	}
	if config.Workflows[0] != "propose" || config.Workflows[1] != "explore" || config.Workflows[2] != "apply" {
		t.Errorf("unexpected workflows: %v", config.Workflows)
	}
	if config.Context != "This is the project context for testing." {
		t.Errorf("unexpected context: %q", config.Context)
	}
	if len(config.Rules) != 2 {
		t.Errorf("expected 2 rule keys, got %d", len(config.Rules))
	}
	if len(config.Rules["proposal"]) != 2 {
		t.Errorf("expected 2 proposal rules, got %d", len(config.Rules["proposal"]))
	}
	if len(config.Rules["spec"]) != 1 {
		t.Errorf("expected 1 spec rule, got %d", len(config.Rules["spec"]))
	}
}

func TestReadProjectConfig_YMLFallback(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: custom-schema
profile: all
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config from .yml fallback")
	}
	if config.Schema != "custom-schema" {
		t.Errorf("expected schema 'custom-schema', got %q", config.Schema)
	}
	if config.Profile != "all" {
		t.Errorf("expected profile 'all', got %q", config.Profile)
	}
}

func TestReadProjectConfig_YAMLPreferredOverYML(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := `schema: from-yaml
`
	ymlContent := `schema: from-yml
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yml"), []byte(ymlContent), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Schema != "from-yaml" {
		t.Errorf("expected .yaml to take precedence, got schema %q", config.Schema)
	}
}

func TestReadProjectConfig_ReturnsNilWhenMissing(t *testing.T) {
	dir := t.TempDir()
	// No openspec directory at all
	config := ReadProjectConfig(dir)
	if config != nil {
		t.Error("expected nil config when no config file exists")
	}
}

func TestReadProjectConfig_ReturnsNilWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config != nil {
		t.Error("expected nil config for empty file")
	}
}

func TestReadProjectConfig_ParsesRules(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
rules:
  proposal:
    - "Rule one for proposal"
    - "Rule two for proposal"
  spec:
    - "Spec rule A"
    - "Spec rule B"
    - "Spec rule C"
  design:
    - "Design rule"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(config.Rules) != 3 {
		t.Fatalf("expected 3 rule keys, got %d", len(config.Rules))
	}
	if len(config.Rules["proposal"]) != 2 {
		t.Errorf("expected 2 proposal rules, got %d", len(config.Rules["proposal"]))
	}
	if len(config.Rules["spec"]) != 3 {
		t.Errorf("expected 3 spec rules, got %d", len(config.Rules["spec"]))
	}
	if len(config.Rules["design"]) != 1 {
		t.Errorf("expected 1 design rule, got %d", len(config.Rules["design"]))
	}
	if config.Rules["proposal"][0] != "Rule one for proposal" {
		t.Errorf("unexpected first proposal rule: %q", config.Rules["proposal"][0])
	}
}

func TestReadProjectConfig_FiltersEmptyStrings(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
rules:
  proposal:
    - "Valid rule"
    - ""
    - "Another valid rule"
    - ""
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(config.Rules["proposal"]) != 2 {
		t.Errorf("expected 2 rules after filtering empty strings, got %d", len(config.Rules["proposal"]))
	}
	for _, rule := range config.Rules["proposal"] {
		if rule == "" {
			t.Error("empty string should have been filtered out")
		}
	}
}

func TestReadProjectConfig_ContextSizeLimit(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create context that exceeds MaxContextSize (50KB)
	largeContext := strings.Repeat("x", MaxContextSize+1)
	content := "schema: spec-driven\ncontext: |\n  " + largeContext + "\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config (schema is still set)")
	}
	if config.Context != "" {
		t.Errorf("expected empty context when exceeding size limit, got %d bytes", len(config.Context))
	}
}

func TestReadProjectConfig_ParsesKeywords(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
keywords:
  normative: ["DEBE", "DEBERA"]
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Keywords == nil {
		t.Fatal("expected non-nil Keywords")
	}
	if len(config.Keywords.Normative) != 2 {
		t.Fatalf("expected 2 normative keywords, got %d", len(config.Keywords.Normative))
	}
	if config.Keywords.Normative[0] != "DEBE" || config.Keywords.Normative[1] != "DEBERA" {
		t.Errorf("unexpected keywords: %v", config.Keywords.Normative)
	}
}

func TestReadProjectConfig_KeywordsAbsent(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Keywords != nil {
		t.Errorf("expected nil Keywords when not in config, got %+v", config.Keywords)
	}
}

func TestReadProjectConfig_KeywordsEmpty(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
keywords:
  normative: []
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Keywords == nil {
		t.Fatal("expected non-nil Keywords even with empty list")
	}
	if len(config.Keywords.Normative) != 0 {
		t.Errorf("expected empty normative list, got %v", config.Keywords.Normative)
	}
}

func TestReadProjectConfig_KeywordsWithAccents(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `schema: spec-driven
keywords:
  normative: ["DEBERÁ", "DEBE"]
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := ReadProjectConfig(dir)
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Keywords == nil {
		t.Fatal("expected non-nil Keywords")
	}
	if len(config.Keywords.Normative) != 2 {
		t.Fatalf("expected 2 normative keywords, got %d", len(config.Keywords.Normative))
	}
	if config.Keywords.Normative[0] != "DEBERÁ" {
		t.Errorf("expected 'DEBERÁ', got %q", config.Keywords.Normative[0])
	}
}

func TestValidateKeywords_Valid(t *testing.T) {
	kw := &KeywordsConfig{Normative: []string{"DEBE", "DEBERA"}}
	warnings := ValidateKeywords(kw)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestValidateKeywords_Nil(t *testing.T) {
	warnings := ValidateKeywords(nil)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for nil, got: %v", warnings)
	}
}

func TestValidateKeywords_Empty(t *testing.T) {
	kw := &KeywordsConfig{Normative: []string{}}
	warnings := ValidateKeywords(kw)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "empty") {
		t.Errorf("expected warning about empty list, got: %s", warnings[0])
	}
}

func TestValidateKeywords_RegexMetachars(t *testing.T) {
	kw := &KeywordsConfig{Normative: []string{"MUST(not)"}}
	warnings := ValidateKeywords(kw)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "regex metacharacters") {
		t.Errorf("expected warning about metacharacters, got: %s", warnings[0])
	}
}

func TestValidateConfigRules_ValidArtifactIDs(t *testing.T) {
	rules := map[string][]string{
		"proposal": {"Rule 1"},
		"spec":     {"Rule 2"},
		"design":   {"Rule 3"},
	}
	validIDs := map[string]bool{
		"proposal": true,
		"spec":     true,
		"design":   true,
		"tasks":    true,
	}

	warnings := ValidateConfigRules(rules, validIDs, "spec-driven")
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid artifact IDs, got: %v", warnings)
	}
}

func TestValidateConfigRules_InvalidArtifactIDs(t *testing.T) {
	rules := map[string][]string{
		"proposal":       {"Rule 1"},
		"nonexistent":    {"Rule 2"},
		"also-not-valid": {"Rule 3"},
	}
	validIDs := map[string]bool{
		"proposal": true,
		"spec":     true,
		"design":   true,
	}

	warnings := ValidateConfigRules(rules, validIDs, "spec-driven")
	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings for invalid artifact IDs, got %d: %v", len(warnings), warnings)
	}

	foundNonexistent := false
	foundAlsoNotValid := false
	for _, w := range warnings {
		if strings.Contains(w, "nonexistent") {
			foundNonexistent = true
		}
		if strings.Contains(w, "also-not-valid") {
			foundAlsoNotValid = true
		}
		if !strings.Contains(w, "spec-driven") {
			t.Errorf("warning should mention schema name, got: %s", w)
		}
	}
	if !foundNonexistent {
		t.Error("expected warning about 'nonexistent' artifact ID")
	}
	if !foundAlsoNotValid {
		t.Error("expected warning about 'also-not-valid' artifact ID")
	}
}
