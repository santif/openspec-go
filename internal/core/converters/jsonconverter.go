package converters

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/santif/openspec-go/internal/core/parsers"
	"github.com/santif/openspec-go/internal/core/schemas"
	"github.com/santif/openspec-go/internal/utils"
)

// extractNameFromPath extracts a spec or change name from its file path.
// It looks for "specs" or "changes" parent directories and returns the next segment.
func extractNameFromPath(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "specs" || parts[i] == "changes" {
			if i < len(parts)-1 {
				return parts[i+1]
			}
		}
	}
	// Fallback: filename without extension
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

// ConvertSpecToJSON reads a spec markdown file and returns its JSON representation.
func ConvertSpecToJSON(filePath string) (string, error) {
	content, err := utils.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read spec file: %w", err)
	}

	specName := extractNameFromPath(filePath)
	spec, err := parsers.ParseSpec(specName, content)
	if err != nil {
		return "", fmt.Errorf("failed to parse spec: %w", err)
	}

	if spec.Metadata == nil {
		spec.Metadata = &schemas.Metadata{}
	}
	spec.Metadata.SourcePath = filePath

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec to JSON: %w", err)
	}

	return string(data), nil
}

// ConvertChangeToJSON reads a change markdown file and returns its JSON representation.
func ConvertChangeToJSON(filePath string) (string, error) {
	content, err := utils.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read change file: %w", err)
	}

	changeName := extractNameFromPath(filePath)
	changeDir := filepath.Dir(filePath)
	change, err := parsers.ParseChangeWithDeltas(changeName, content, changeDir)
	if err != nil {
		return "", fmt.Errorf("failed to parse change: %w", err)
	}

	if change.Metadata == nil {
		change.Metadata = &schemas.Metadata{}
	}
	change.Metadata.SourcePath = filePath

	data, err := json.MarshalIndent(change, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal change to JSON: %w", err)
	}

	return string(data), nil
}
