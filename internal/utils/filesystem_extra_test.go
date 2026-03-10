package utils

import (
	"strings"
	"testing"
)

func TestToPosixPath(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"backslashes", `a\b\c`, "a/b/c"},
		{"mixed separators", `a/b\c`, "a/b/c"},
		{"already posix", "a/b/c", "a/b/c"},
		{"empty string", "", ""},
		{"single backslash", `\`, "/"},
		{"windows path", `C:\Users\test\file.txt`, "C:/Users/test/file.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToPosixPath(tc.input)
			if got != tc.expected {
				t.Errorf("ToPosixPath(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestRemoveMarkerBlock_NotFound(t *testing.T) {
	content := "line1\nline2\nline3\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if got != content {
		t.Errorf("expected unchanged content when markers not found, got %q", got)
	}
}

func TestRemoveMarkerBlock_BasicRemoval(t *testing.T) {
	content := "before\n<!-- START -->\ninner content\n<!-- END -->\nafter\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if strings.Contains(got, "inner content") {
		t.Error("expected inner content to be removed")
	}
	if strings.Contains(got, "<!-- START -->") {
		t.Error("expected start marker to be removed")
	}
	if strings.Contains(got, "<!-- END -->") {
		t.Error("expected end marker to be removed")
	}
	if !strings.Contains(got, "before") {
		t.Error("expected 'before' to be preserved")
	}
	if !strings.Contains(got, "after") {
		t.Error("expected 'after' to be preserved")
	}
}

func TestRemoveMarkerBlock_AtStartOfFile(t *testing.T) {
	content := "<!-- START -->\ninner\n<!-- END -->\nafter\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if strings.Contains(got, "inner") {
		t.Error("expected inner content to be removed")
	}
	if !strings.Contains(got, "after") {
		t.Error("expected 'after' to be preserved")
	}
}

func TestRemoveMarkerBlock_OnlyMarkers(t *testing.T) {
	content := "<!-- START -->\ninner\n<!-- END -->\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if got != "" {
		t.Errorf("expected empty string when only markers exist, got %q", got)
	}
}

func TestRemoveMarkerBlock_TripleBlanksNormalized(t *testing.T) {
	content := "before\n\n\n<!-- START -->\ninner\n<!-- END -->\n\n\nafter\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if strings.Contains(got, "\n\n\n") {
		t.Error("expected triple blank lines to be normalized")
	}
}

func TestRemoveMarkerBlock_EndBeforeStart(t *testing.T) {
	// If end marker appears before start marker in content, should return unchanged
	content := "<!-- END -->\nstuff\n<!-- START -->\n"
	got := RemoveMarkerBlock(content, "<!-- START -->", "<!-- END -->")
	if got != content {
		t.Errorf("expected unchanged content when end is before start, got %q", got)
	}
}

func TestIsMarkerOnOwnLine(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		index    int
		length   int
		expected bool
	}{
		{"marker at start of file", "MARKER\nrest", 0, 6, true},
		{"marker at end of file", "rest\nMARKER", 5, 6, true},
		{"marker with leading spaces", "  MARKER\nrest", 2, 6, true},
		{"marker with trailing spaces", "MARKER  \nrest", 0, 6, true},
		{"marker with leading tab", "\tMARKER\nrest", 1, 6, true},
		{"marker with text before", "textMARKER\nrest", 4, 6, false},
		{"marker with text after", "MARKERtext\nrest", 0, 6, false},
		{"marker on middle line", "before\nMARKER\nafter", 7, 6, true},
		{"marker in middle of text line", "before MARKER after\n", 7, 6, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isMarkerOnOwnLine(tc.content, tc.index, tc.length)
			if got != tc.expected {
				t.Errorf("isMarkerOnOwnLine(%q, %d, %d) = %v, want %v",
					tc.content, tc.index, tc.length, got, tc.expected)
			}
		})
	}
}
