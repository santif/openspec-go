package artifactgraph

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// DetectCompleted checks the filesystem to determine which artifacts have been generated.
func DetectCompleted(graph *ArtifactGraph, changeDir string) map[string]bool {
	completed := make(map[string]bool)

	for _, artifact := range graph.GetAllArtifacts() {
		pattern := artifact.Generates
		// Check if the glob pattern matches any file in changeDir
		fullPattern := filepath.Join(changeDir, pattern)

		// Use doublestar for ** glob support
		matches, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			continue
		}

		if len(matches) > 0 {
			// Verify at least one match is a non-empty file
			for _, match := range matches {
				info, err := os.Stat(match)
				if err == nil && !info.IsDir() && info.Size() > 0 {
					completed[artifact.ID] = true
					break
				}
			}
		}
	}

	return completed
}
