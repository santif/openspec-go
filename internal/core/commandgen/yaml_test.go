package commandgen

import "testing"

func TestEscapeYamlValue(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"simple string", "hello", "hello"},
		{"with colon", "key: value", `"key: value"`},
		{"with newline", "line1\nline2", `"line1\nline2"`},
		{"with hash", "# comment", `"# comment"`},
		{"with braces", "{}", `"{}"`},
		{"with double quote", `say "hi"`, `"say \"hi\""`},
		{"with backslash", `a\b`, `a\b`}, // backslash not in YAML special chars regex
		{"leading space", " leading", `" leading"`},
		{"trailing space", "trailing ", `"trailing "`},
		{"empty string", "", ""},
		{"with ampersand", "&ref", `"&ref"`},
		{"with asterisk", "*alias", `"*alias"`},
		{"with exclamation", "!tag", `"!tag"`},
		{"with pipe", "a|b", `"a|b"`},
		{"with percent", "100%", `"100%"`}, // % is in YAML special chars regex
		{"with square bracket", "[item]", `"[item]"`},
		{"with comma", "a, b", `"a, b"`},
		{"with backtick", "code `here`", "\"code `here`\""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeYamlValue(tc.input)
			if got != tc.expected {
				t.Errorf("EscapeYamlValue(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestFormatTagsArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty", nil, ""},
		{"single", []string{"tag1"}, "\n  - tag1"},
		{"multiple", []string{"tag1", "tag2", "tag3"}, "\n  - tag1\n  - tag2\n  - tag3"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatTagsArray(tc.input)
			if got != tc.expected {
				t.Errorf("FormatTagsArray(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
