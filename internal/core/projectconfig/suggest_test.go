package projectconfig

import (
	"strings"
	"testing"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		expected int
	}{
		{"both empty", "", "", 0},
		{"first empty", "", "abc", 3},
		{"second empty", "abc", "", 3},
		{"identical", "abc", "abc", 0},
		{"kitten-sitting", "kitten", "sitting", 3},
		{"single substitution", "abc", "abd", 1},
		{"single char different", "a", "b", 1},
		{"single char same", "a", "a", 0},
		{"insertion", "abc", "abcd", 1},
		{"deletion", "abcd", "abc", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := levenshtein(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}

func TestJoinList(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty", nil, ""},
		{"single", []string{"a"}, "a"},
		{"two", []string{"a", "b"}, "a, b"},
		{"three", []string{"a", "b", "c"}, "a, b, c"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := joinList(tc.input)
			if got != tc.expected {
				t.Errorf("joinList(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSuggestSchemas_CloseMatch(t *testing.T) {
	schemas := []AvailableSchema{
		{Name: "spec-driven", IsBuiltIn: true},
		{Name: "custom-flow", IsBuiltIn: false},
	}
	result := SuggestSchemas("spec-drven", schemas)
	if !strings.Contains(result, "Did you mean") {
		t.Error("expected suggestions section")
	}
	if !strings.Contains(result, "spec-driven") {
		t.Error("expected spec-driven in suggestions")
	}
}

func TestSuggestSchemas_NoCloseMatch(t *testing.T) {
	schemas := []AvailableSchema{
		{Name: "completely-different", IsBuiltIn: true},
	}
	result := SuggestSchemas("xyz", schemas)
	if strings.Contains(result, "Did you mean") {
		t.Error("expected no suggestions for distant names")
	}
}

func TestSuggestSchemas_MaxThreeSuggestions(t *testing.T) {
	schemas := []AvailableSchema{
		{Name: "ab", IsBuiltIn: true},
		{Name: "ac", IsBuiltIn: true},
		{Name: "ad", IsBuiltIn: true},
		{Name: "ae", IsBuiltIn: true},
	}
	result := SuggestSchemas("aa", schemas)
	// Count suggestion lines (format: "  - name (type)")
	count := strings.Count(result, "  - ")
	// At most 3 in the "Did you mean" section + available section entries
	if strings.Contains(result, "Did you mean") {
		// Extract just the suggestion part between "Did you mean" and "Available"
		parts := strings.SplitN(result, "Available schemas:", 2)
		suggestionPart := parts[0]
		suggestionLines := strings.Count(suggestionPart, "  - ")
		if suggestionLines > 3 {
			t.Errorf("expected at most 3 suggestions, got %d (count=%d)", suggestionLines, count)
		}
	}
}

func TestSuggestSchemas_MixedBuiltInAndProjectLocal(t *testing.T) {
	schemas := []AvailableSchema{
		{Name: "spec-driven", IsBuiltIn: true},
		{Name: "my-custom", IsBuiltIn: false},
	}
	result := SuggestSchemas("nonexistent-schema", schemas)
	if !strings.Contains(result, "Built-in:") {
		t.Error("expected Built-in section")
	}
	if !strings.Contains(result, "Project-local:") {
		t.Error("expected Project-local section")
	}
}

func TestSuggestSchemas_EmptyAvailable(t *testing.T) {
	result := SuggestSchemas("anything", nil)
	if !strings.Contains(result, "(none found)") {
		t.Error("expected '(none found)' for empty project-local schemas")
	}
}

func TestSuggestSchemas_ContainsFixHint(t *testing.T) {
	result := SuggestSchemas("bad-name", nil)
	if !strings.Contains(result, "Fix:") {
		t.Error("expected Fix hint in output")
	}
	if !strings.Contains(result, "bad-name") {
		t.Error("expected invalid name in Fix hint")
	}
}
