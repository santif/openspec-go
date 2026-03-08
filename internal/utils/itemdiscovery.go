package utils

import (
	"os"
	"path/filepath"
	"sort"
)

// GetActiveChangeIDs returns sorted directory names under openspec/changes/.
func GetActiveChangeIDs(projectRoot string) []string {
	changesDir := filepath.Join(projectRoot, "openspec", "changes")
	return listSubdirectories(changesDir)
}

// GetSpecIDs returns sorted directory names under openspec/specs/.
func GetSpecIDs(projectRoot string) []string {
	specsDir := filepath.Join(projectRoot, "openspec", "specs")
	return listSubdirectories(specsDir)
}

// GetArchivedChangeIDs returns sorted directory names under openspec/archive/.
func GetArchivedChangeIDs(projectRoot string) []string {
	archiveDir := filepath.Join(projectRoot, "openspec", "archive")
	return listSubdirectories(archiveDir)
}

func listSubdirectories(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var ids []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 0 && entry.Name()[0] != '.' {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	return ids
}
