package commandgen

import "strings"

// TransformToHyphenCommands replaces colon-based command references
// (e.g., /opsx:propose) with hyphen-based format (e.g., /opsx-propose).
// This is used by tools like OpenCode that use hyphens instead of colons.
func TransformToHyphenCommands(content string) string {
	return strings.ReplaceAll(content, "/opsx:", "/opsx-")
}
