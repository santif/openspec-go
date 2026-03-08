package artifactgraph

import (
	"fmt"
	"strings"
)

// EnrichedInstruction holds the combined instruction text for an artifact.
type EnrichedInstruction struct {
	ArtifactID  string `json:"artifactId"`
	Instruction string `json:"instruction"`
}

// LoadEnrichedInstruction builds an enriched instruction for an artifact,
// combining the base instruction with project context and rules.
func LoadEnrichedInstruction(graph *ArtifactGraph, artifactID string, context string, rules map[string][]string) (*EnrichedInstruction, error) {
	artifact := graph.GetArtifact(artifactID)
	if artifact == nil {
		return nil, fmt.Errorf("artifact %q not found in schema", artifactID)
	}

	var parts []string

	// Base instruction from schema
	if artifact.Instruction != "" {
		parts = append(parts, artifact.Instruction)
	}

	// Project context
	if context != "" {
		parts = append(parts, fmt.Sprintf("<project-context>\n%s\n</project-context>", context))
	}

	// Per-artifact rules
	if rules != nil {
		if artifactRules, ok := rules[artifactID]; ok && len(artifactRules) > 0 {
			rulesText := "<rules>\n"
			for _, rule := range artifactRules {
				rulesText += fmt.Sprintf("- %s\n", rule)
			}
			rulesText += "</rules>"
			parts = append(parts, rulesText)
		}
	}

	return &EnrichedInstruction{
		ArtifactID:  artifactID,
		Instruction: strings.Join(parts, "\n\n"),
	}, nil
}

// LoadApplyInstruction builds an enriched instruction for the apply phase.
func LoadApplyInstruction(graph *ArtifactGraph, context string, rules map[string][]string) (*EnrichedInstruction, error) {
	schema := graph.GetSchema()
	if schema.Apply == nil {
		return nil, fmt.Errorf("schema %q does not define an apply phase", schema.Name)
	}

	var parts []string

	if schema.Apply.Instruction != "" {
		parts = append(parts, schema.Apply.Instruction)
	}

	if context != "" {
		parts = append(parts, fmt.Sprintf("<project-context>\n%s\n</project-context>", context))
	}

	// Apply-specific rules
	if rules != nil {
		if applyRules, ok := rules["apply"]; ok && len(applyRules) > 0 {
			rulesText := "<rules>\n"
			for _, rule := range applyRules {
				rulesText += fmt.Sprintf("- %s\n", rule)
			}
			rulesText += "</rules>"
			parts = append(parts, rulesText)
		}
	}

	return &EnrichedInstruction{
		ArtifactID:  "apply",
		Instruction: strings.Join(parts, "\n\n"),
	}, nil
}
