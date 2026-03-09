package commandgen

import (
	"regexp"
	"strings"
)

// yamlSpecialChars matches characters that require YAML value quoting.
var yamlSpecialChars = regexp.MustCompile(`[:\n\r#{}[\],&*!|>'"%@` + "`" + `]|^\s|\s$`)

// EscapeYamlValue returns a safely-quoted YAML value string.
// If the value contains special YAML characters, it is double-quoted with
// internal backslashes, double-quotes, and newlines escaped.
func EscapeYamlValue(value string) string {
	if yamlSpecialChars.MatchString(value) {
		escaped := strings.ReplaceAll(value, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		return `"` + escaped + `"`
	}
	return value
}

// FormatTagsArray formats a string slice as a YAML array block.
func FormatTagsArray(tags []string) string {
	var b strings.Builder
	for _, tag := range tags {
		b.WriteString("\n  - ")
		b.WriteString(tag)
	}
	return b.String()
}
