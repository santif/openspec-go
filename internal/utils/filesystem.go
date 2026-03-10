package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ToPosixPath(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func UpdateFileWithMarkers(filePath, content, startMarker, endMarker string) error {
	existingContent := ""
	if FileExists(filePath) {
		data, err := ReadFile(filePath)
		if err != nil {
			return err
		}
		existingContent = data

		startIndex := findMarkerIndex(existingContent, startMarker, 0)
		var endIndex int
		if startIndex != -1 {
			endIndex = findMarkerIndex(existingContent, endMarker, startIndex+len(startMarker))
		} else {
			endIndex = findMarkerIndex(existingContent, endMarker, 0)
		}

		if startIndex != -1 && endIndex != -1 {
			if endIndex < startIndex {
				return fmt.Errorf("invalid marker state in %s: end marker appears before start marker", filePath)
			}
			before := existingContent[:startIndex]
			after := existingContent[endIndex+len(endMarker):]
			existingContent = before + startMarker + "\n" + content + "\n" + endMarker + after
		} else if startIndex == -1 && endIndex == -1 {
			existingContent = startMarker + "\n" + content + "\n" + endMarker + "\n\n" + existingContent
		} else {
			return fmt.Errorf("invalid marker state in %s: found start: %v, found end: %v", filePath, startIndex != -1, endIndex != -1)
		}
	} else {
		existingContent = startMarker + "\n" + content + "\n" + endMarker
	}

	return WriteFile(filePath, existingContent)
}

// RemoveMarkerBlock removes a marker block (start marker through end marker,
// inclusive of the lines they appear on) from content. Returns the original
// content unchanged if markers are not found or invalid.
func RemoveMarkerBlock(content, startMarker, endMarker string) string {
	startIndex := findMarkerIndex(content, startMarker, 0)
	var endIndex int
	if startIndex != -1 {
		endIndex = findMarkerIndex(content, endMarker, startIndex+len(startMarker))
	} else {
		endIndex = findMarkerIndex(content, endMarker, 0)
	}

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return content
	}

	// Find the start of the line containing the start marker
	lineStart := startIndex
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Find the end of the line containing the end marker
	lineEnd := endIndex + len(endMarker)
	for lineEnd < len(content) && content[lineEnd] != '\n' {
		lineEnd++
	}
	// Include the trailing newline if present
	if lineEnd < len(content) && content[lineEnd] == '\n' {
		lineEnd++
	}

	before := content[:lineStart]
	after := content[lineEnd:]

	result := before + after

	// Clean up triple+ blank lines
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}

	trimmed := strings.TrimRight(result, " \t\n\r")
	if trimmed == "" {
		return ""
	}
	return trimmed + "\n"
}

func findMarkerIndex(content, marker string, fromIndex int) int {
	if fromIndex >= len(content) {
		return -1
	}
	idx := strings.Index(content[fromIndex:], marker)
	if idx == -1 {
		return -1
	}
	absIdx := fromIndex + idx
	if isMarkerOnOwnLine(content, absIdx, len(marker)) {
		return absIdx
	}
	// Search for next occurrence
	return findMarkerIndex(content, marker, absIdx+len(marker))
}

func isMarkerOnOwnLine(content string, markerIndex, markerLength int) bool {
	// Check left side
	leftIndex := markerIndex - 1
	for leftIndex >= 0 && content[leftIndex] != '\n' {
		ch := content[leftIndex]
		if ch != ' ' && ch != '\t' && ch != '\r' {
			return false
		}
		leftIndex--
	}
	// Check right side
	rightIndex := markerIndex + markerLength
	for rightIndex < len(content) && content[rightIndex] != '\n' {
		ch := content[rightIndex]
		if ch != ' ' && ch != '\t' && ch != '\r' {
			return false
		}
		rightIndex++
	}
	return true
}
