package projectconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

var regexMetaChars = regexp.MustCompile(`[.*+?^${}()|[\]\\]`)

// MaxContextSize is the maximum allowed size for the context field (50KB).
const MaxContextSize = 50 * 1024

// ConditionalsConfig holds configurable conditional keyword settings.
type ConditionalsConfig struct {
	When string `yaml:"when" json:"when"`
	Then string `yaml:"then" json:"then"`
	And  string `yaml:"and" json:"and"`
}

// DefaultConditionals returns the default conditional keywords.
func DefaultConditionals() ConditionalsConfig {
	return ConditionalsConfig{When: "WHEN", Then: "THEN", And: "AND"}
}

// ResolveConditionals returns the effective conditional keywords.
// If Keywords or Conditionals is nil, returns defaults.
func ResolveConditionals(kw *KeywordsConfig) ConditionalsConfig {
	if kw == nil || kw.Conditionals == nil {
		return DefaultConditionals()
	}
	return *kw.Conditionals
}

// KeywordsConfig holds configurable keyword settings.
type KeywordsConfig struct {
	Normative    []string            `yaml:"normative" json:"normative"`
	Conditionals *ConditionalsConfig `yaml:"conditionals,omitempty" json:"conditionals,omitempty"`
}

// ProjectConfig represents the openspec/config.yaml file.
type ProjectConfig struct {
	Schema    string              `yaml:"schema" json:"schema"`
	Profile   string              `yaml:"profile,omitempty" json:"profile,omitempty"`
	Workflows []string            `yaml:"workflows,omitempty" json:"workflows,omitempty"`
	Context   string              `yaml:"context,omitempty" json:"context,omitempty"`
	Rules     map[string][]string `yaml:"rules,omitempty" json:"rules,omitempty"`
	Keywords  *KeywordsConfig     `yaml:"keywords,omitempty" json:"keywords,omitempty"`
}

// ReadProjectConfig reads and parses openspec/config.yaml from the project root.
// Returns nil if the config file does not exist or is empty.
func ReadProjectConfig(projectRoot string) *ProjectConfig {
	configPath := filepath.Join(projectRoot, "openspec", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		configPath = filepath.Join(projectRoot, "openspec", "config.yml")
		if _, err := os.Stat(configPath); err != nil {
			return nil
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to read %s: %v\n", configPath, err)
		return nil
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: openspec/config.yaml is not valid YAML: %v\n", err)
		return nil
	}

	config := &ProjectConfig{}

	// Parse schema
	if s, ok := raw["schema"].(string); ok && s != "" {
		config.Schema = s
	}

	// Parse profile
	if p, ok := raw["profile"].(string); ok && p != "" {
		config.Profile = p
	}

	// Parse workflows
	if w, ok := raw["workflows"].([]interface{}); ok {
		for _, item := range w {
			if s, ok := item.(string); ok && s != "" {
				config.Workflows = append(config.Workflows, s)
			}
		}
	}

	// Parse context with size limit
	if c, ok := raw["context"].(string); ok {
		if len(c) > MaxContextSize {
			fmt.Fprintf(os.Stderr, "Warning: Context too large (%dKB, limit: %dKB), ignoring\n", len(c)/1024, MaxContextSize/1024)
		} else {
			config.Context = c
		}
	}

	// Parse rules
	if r, ok := raw["rules"].(map[string]interface{}); ok {
		config.Rules = make(map[string][]string)
		for artifactID, rulesRaw := range r {
			if rulesList, ok := rulesRaw.([]interface{}); ok {
				var rules []string
				for _, rule := range rulesList {
					if s, ok := rule.(string); ok && s != "" {
						rules = append(rules, s)
					}
				}
				if len(rules) > 0 {
					config.Rules[artifactID] = rules
				}
			}
		}
	}

	// Parse keywords
	if kw, ok := raw["keywords"].(map[string]interface{}); ok {
		kwConfig := &KeywordsConfig{}
		if normative, ok := kw["normative"].([]interface{}); ok {
			for _, item := range normative {
				if s, ok := item.(string); ok {
					kwConfig.Normative = append(kwConfig.Normative, s)
				}
			}
		}
		if cond, ok := kw["conditionals"].(map[string]interface{}); ok {
			condConfig := &ConditionalsConfig{}
			if w, ok := cond["when"].(string); ok {
				condConfig.When = w
			}
			if t, ok := cond["then"].(string); ok {
				condConfig.Then = t
			}
			if a, ok := cond["and"].(string); ok {
				condConfig.And = a
			}
			kwConfig.Conditionals = condConfig
		}
		config.Keywords = kwConfig
	}

	if config.Schema == "" && config.Context == "" && config.Profile == "" &&
		len(config.Rules) == 0 && len(config.Workflows) == 0 && config.Keywords == nil {
		return nil
	}

	return config
}

// ValidateKeywords checks the keywords configuration for issues.
// Returns warnings (not errors) so projects can still function.
func ValidateKeywords(kw *KeywordsConfig) []string {
	if kw == nil {
		return nil
	}
	var warnings []string
	if len(kw.Normative) == 0 {
		warnings = append(warnings, "keywords.normative is empty; validator will fall back to default keywords (SHALL, MUST)")
	}
	for _, keyword := range kw.Normative {
		if keyword == "" {
			warnings = append(warnings, "keywords.normative contains an empty string")
			continue
		}
		if regexMetaChars.MatchString(keyword) {
			warnings = append(warnings, fmt.Sprintf("keyword %q contains regex metacharacters; this may cause unexpected behavior", keyword))
		}
	}
	if kw.Conditionals != nil {
		condFields := map[string]string{
			"when": kw.Conditionals.When,
			"then": kw.Conditionals.Then,
			"and":  kw.Conditionals.And,
		}
		for name, value := range condFields {
			if value == "" {
				warnings = append(warnings, fmt.Sprintf("keywords.conditionals.%s is empty", name))
			} else if regexMetaChars.MatchString(value) {
				warnings = append(warnings, fmt.Sprintf("conditional keyword %q for %q contains regex metacharacters; this may cause unexpected behavior", value, name))
			}
		}
	}
	return warnings
}

// ValidateConfigRules checks that rule keys reference valid artifact IDs.
func ValidateConfigRules(rules map[string][]string, validArtifactIDs map[string]bool, schemaName string) []string {
	var warnings []string
	for artifactID := range rules {
		if !validArtifactIDs[artifactID] {
			var validIDs []string
			for id := range validArtifactIDs {
				validIDs = append(validIDs, id)
			}
			sort.Strings(validIDs)
			warnings = append(warnings, fmt.Sprintf(
				"Unknown artifact ID in rules: %q. Valid IDs for schema %q: %s",
				artifactID, schemaName, joinSorted(validIDs),
			))
		}
	}
	return warnings
}

func joinSorted(items []string) string {
	sorted := make([]string, len(items))
	copy(sorted, items)
	sort.Strings(sorted)
	result := ""
	for i, s := range sorted {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
