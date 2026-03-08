package artifactgraph

import "sort"

// ArtifactGraph provides graph operations over a schema's artifacts.
type ArtifactGraph struct {
	artifacts map[string]*Artifact
	schema    *SchemaYaml
}

// NewGraphFromYaml loads a schema file and constructs a graph.
func NewGraphFromYaml(filePath string) (*ArtifactGraph, error) {
	schema, err := LoadSchema(filePath)
	if err != nil {
		return nil, err
	}
	return NewGraphFromSchema(schema), nil
}

// NewGraphFromYamlContent parses schema YAML content and constructs a graph.
func NewGraphFromYamlContent(yamlContent string) (*ArtifactGraph, error) {
	schema, err := ParseSchema(yamlContent)
	if err != nil {
		return nil, err
	}
	return NewGraphFromSchema(schema), nil
}

// NewGraphFromSchema constructs a graph from a parsed schema.
func NewGraphFromSchema(schema *SchemaYaml) *ArtifactGraph {
	artifacts := make(map[string]*Artifact)
	for i := range schema.Artifacts {
		artifacts[schema.Artifacts[i].ID] = &schema.Artifacts[i]
	}
	return &ArtifactGraph{
		artifacts: artifacts,
		schema:    schema,
	}
}

// GetArtifact returns the artifact with the given ID, or nil if not found.
func (g *ArtifactGraph) GetArtifact(id string) *Artifact {
	return g.artifacts[id]
}

// GetAllArtifacts returns all artifacts in schema order.
func (g *ArtifactGraph) GetAllArtifacts() []Artifact {
	result := make([]Artifact, 0, len(g.artifacts))
	for _, a := range g.schema.Artifacts {
		result = append(result, a)
	}
	return result
}

// GetName returns the schema name.
func (g *ArtifactGraph) GetName() string {
	return g.schema.Name
}

// GetVersion returns the schema version.
func (g *ArtifactGraph) GetVersion() int {
	return g.schema.Version
}

// GetSchema returns the underlying schema.
func (g *ArtifactGraph) GetSchema() *SchemaYaml {
	return g.schema
}

// GetBuildOrder returns artifact IDs in topological order using Kahn's algorithm.
func (g *ArtifactGraph) GetBuildOrder() []string {
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for _, a := range g.schema.Artifacts {
		inDegree[a.ID] = len(a.Requires)
		dependents[a.ID] = nil
	}

	for _, a := range g.schema.Artifacts {
		for _, req := range a.Requires {
			dependents[req] = append(dependents[req], a.ID)
		}
	}

	// Start with roots (in-degree 0), sorted for determinism
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		var newlyReady []string
		for _, dep := range dependents[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				newlyReady = append(newlyReady, dep)
			}
		}
		sort.Strings(newlyReady)
		queue = append(queue, newlyReady...)
	}

	return result
}

// GetNextArtifacts returns artifact IDs whose dependencies are all in the completed set.
func (g *ArtifactGraph) GetNextArtifacts(completed map[string]bool) []string {
	var ready []string
	for _, a := range g.schema.Artifacts {
		if completed[a.ID] {
			continue
		}
		allDepsCompleted := true
		for _, req := range a.Requires {
			if !completed[req] {
				allDepsCompleted = false
				break
			}
		}
		if allDepsCompleted {
			ready = append(ready, a.ID)
		}
	}
	sort.Strings(ready)
	return ready
}

// IsComplete returns true if all artifacts are in the completed set.
func (g *ArtifactGraph) IsComplete(completed map[string]bool) bool {
	for _, a := range g.schema.Artifacts {
		if !completed[a.ID] {
			return false
		}
	}
	return true
}

// GetBlocked returns artifacts that have unmet dependencies, mapping each to its missing deps.
func (g *ArtifactGraph) GetBlocked(completed map[string]bool) BlockedArtifacts {
	blocked := make(BlockedArtifacts)
	for _, a := range g.schema.Artifacts {
		if completed[a.ID] {
			continue
		}
		var unmetDeps []string
		for _, req := range a.Requires {
			if !completed[req] {
				unmetDeps = append(unmetDeps, req)
			}
		}
		if len(unmetDeps) > 0 {
			sort.Strings(unmetDeps)
			blocked[a.ID] = unmetDeps
		}
	}
	return blocked
}
