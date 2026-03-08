package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ChangeMetadata represents the .openspec.yaml file in a change directory.
type ChangeMetadata struct {
	Schema  string `yaml:"schema"`
	Created string `yaml:"created,omitempty"`
}

// ReadChangeMetadata reads and parses .openspec.yaml from the given change directory.
func ReadChangeMetadata(changeDir string) (*ChangeMetadata, error) {
	metaPath := filepath.Join(changeDir, ".openspec.yaml")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta ChangeMetadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// WriteChangeMetadata writes .openspec.yaml to the given change directory.
func WriteChangeMetadata(changeDir string, meta *ChangeMetadata) error {
	metaPath := filepath.Join(changeDir, ".openspec.yaml")
	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	return WriteFile(metaPath, string(data))
}
