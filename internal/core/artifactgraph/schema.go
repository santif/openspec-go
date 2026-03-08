package artifactgraph

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SchemaValidationError represents a schema validation failure.
type SchemaValidationError struct {
	Message string
}

func (e *SchemaValidationError) Error() string {
	return e.Message
}

// LoadSchema reads and parses a schema YAML file.
func LoadSchema(filePath string) (*SchemaYaml, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}
	return ParseSchema(string(content))
}

// ParseSchema parses and validates a schema from YAML content.
func ParseSchema(yamlContent string) (*SchemaYaml, error) {
	var schema SchemaYaml
	if err := yaml.Unmarshal([]byte(yamlContent), &schema); err != nil {
		return nil, &SchemaValidationError{Message: fmt.Sprintf("Invalid schema YAML: %v", err)}
	}

	// Validate required fields
	if schema.Name == "" {
		return nil, &SchemaValidationError{Message: "Invalid schema: name is required"}
	}
	if schema.Version <= 0 {
		return nil, &SchemaValidationError{Message: "Invalid schema: version must be a positive integer"}
	}
	if len(schema.Artifacts) == 0 {
		return nil, &SchemaValidationError{Message: "Invalid schema: at least one artifact required"}
	}
	for i, a := range schema.Artifacts {
		if a.ID == "" {
			return nil, &SchemaValidationError{Message: fmt.Sprintf("Invalid schema: artifacts[%d].id is required", i)}
		}
		if a.Generates == "" {
			return nil, &SchemaValidationError{Message: fmt.Sprintf("Invalid schema: artifacts[%d].generates is required", i)}
		}
		if a.Template == "" {
			return nil, &SchemaValidationError{Message: fmt.Sprintf("Invalid schema: artifacts[%d].template is required", i)}
		}
		// Initialize nil requires to empty slice
		if schema.Artifacts[i].Requires == nil {
			schema.Artifacts[i].Requires = []string{}
		}
	}

	if err := validateNoDuplicateIDs(schema.Artifacts); err != nil {
		return nil, err
	}
	if err := validateRequiresReferences(schema.Artifacts); err != nil {
		return nil, err
	}
	if err := validateNoCycles(schema.Artifacts); err != nil {
		return nil, err
	}

	return &schema, nil
}

func validateNoDuplicateIDs(artifacts []Artifact) error {
	seen := make(map[string]bool)
	for _, a := range artifacts {
		if seen[a.ID] {
			return &SchemaValidationError{Message: fmt.Sprintf("Duplicate artifact ID: %s", a.ID)}
		}
		seen[a.ID] = true
	}
	return nil
}

func validateRequiresReferences(artifacts []Artifact) error {
	validIDs := make(map[string]bool)
	for _, a := range artifacts {
		validIDs[a.ID] = true
	}
	for _, a := range artifacts {
		for _, req := range a.Requires {
			if !validIDs[req] {
				return &SchemaValidationError{
					Message: fmt.Sprintf("Invalid dependency reference in artifact '%s': '%s' does not exist", a.ID, req),
				}
			}
		}
	}
	return nil
}

func validateNoCycles(artifacts []Artifact) error {
	artifactMap := make(map[string]*Artifact)
	for i := range artifacts {
		artifactMap[artifacts[i].ID] = &artifacts[i]
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	parent := make(map[string]string)

	var dfs func(id string) string
	dfs = func(id string) string {
		visited[id] = true
		inStack[id] = true

		a, ok := artifactMap[id]
		if !ok {
			return ""
		}

		for _, dep := range a.Requires {
			if !visited[dep] {
				parent[dep] = id
				if cycle := dfs(dep); cycle != "" {
					return cycle
				}
			} else if inStack[dep] {
				// Found cycle - reconstruct path
				cyclePath := []string{dep}
				current := id
				for current != dep {
					cyclePath = append([]string{current}, cyclePath...)
					current = parent[current]
				}
				cyclePath = append([]string{dep}, cyclePath...)
				return strings.Join(cyclePath, " → ")
			}
		}

		inStack[id] = false
		return ""
	}

	for _, a := range artifacts {
		if !visited[a.ID] {
			if cycle := dfs(a.ID); cycle != "" {
				return &SchemaValidationError{
					Message: fmt.Sprintf("Cyclic dependency detected: %s", cycle),
				}
			}
		}
	}

	return nil
}
