package artifactgraph

// Artifact represents a single artifact in a schema.
type Artifact struct {
	ID          string   `yaml:"id" json:"id"`
	Generates   string   `yaml:"generates" json:"generates"`
	Description string   `yaml:"description" json:"description"`
	Template    string   `yaml:"template" json:"template"`
	Instruction string   `yaml:"instruction,omitempty" json:"instruction,omitempty"`
	Requires    []string `yaml:"requires" json:"requires"`
}

// ApplyPhase defines the optional apply phase of a schema.
type ApplyPhase struct {
	Requires    []string `yaml:"requires" json:"requires"`
	Tracks      string   `yaml:"tracks,omitempty" json:"tracks,omitempty"`
	Instruction string   `yaml:"instruction,omitempty" json:"instruction,omitempty"`
}

// SchemaYaml represents the parsed schema YAML file.
type SchemaYaml struct {
	Name        string      `yaml:"name" json:"name"`
	Version     int         `yaml:"version" json:"version"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Artifacts   []Artifact  `yaml:"artifacts" json:"artifacts"`
	Apply       *ApplyPhase `yaml:"apply,omitempty" json:"apply,omitempty"`
}

// ChangeMetadata holds metadata for a change directory (.openspec.yaml).
type ChangeMetadata struct {
	Schema  string `yaml:"schema" json:"schema"`
	Created string `yaml:"created,omitempty" json:"created,omitempty"`
}

// BlockedArtifacts maps artifact IDs to their unmet dependency IDs.
type BlockedArtifacts map[string][]string
