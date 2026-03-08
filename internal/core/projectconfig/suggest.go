package projectconfig

import "fmt"

// AvailableSchema describes a schema available for use.
type AvailableSchema struct {
	Name      string
	IsBuiltIn bool
}

// SuggestSchemas returns a helpful error message when a schema name is invalid,
// including suggestions based on Levenshtein distance.
func SuggestSchemas(invalidSchemaName string, availableSchemas []AvailableSchema) string {
	type scored struct {
		AvailableSchema
		distance int
	}

	var suggestions []scored
	for _, s := range availableSchemas {
		d := levenshtein(invalidSchemaName, s.Name)
		if d <= 3 {
			suggestions = append(suggestions, scored{s, d})
		}
	}
	// Sort by distance
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].distance < suggestions[i].distance {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	var builtIn, projectLocal []string
	for _, s := range availableSchemas {
		if s.IsBuiltIn {
			builtIn = append(builtIn, s.Name)
		} else {
			projectLocal = append(projectLocal, s.Name)
		}
	}

	msg := fmt.Sprintf("Schema '%s' not found in openspec/config.yaml\n\n", invalidSchemaName)

	if len(suggestions) > 0 {
		msg += "Did you mean one of these?\n"
		for _, s := range suggestions {
			typ := "project-local"
			if s.IsBuiltIn {
				typ = "built-in"
			}
			msg += fmt.Sprintf("  - %s (%s)\n", s.Name, typ)
		}
		msg += "\n"
	}

	msg += "Available schemas:\n"
	if len(builtIn) > 0 {
		msg += fmt.Sprintf("  Built-in: %s\n", joinList(builtIn))
	}
	if len(projectLocal) > 0 {
		msg += fmt.Sprintf("  Project-local: %s\n", joinList(projectLocal))
	} else {
		msg += "  Project-local: (none found)\n"
	}

	msg += fmt.Sprintf("\nFix: Edit openspec/config.yaml and change 'schema: %s' to a valid schema name", invalidSchemaName)

	return msg
}

func joinList(items []string) string {
	result := ""
	for i, s := range items {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	matrix := make([][]int, lb+1)
	for i := range matrix {
		matrix[i] = make([]int, la+1)
		matrix[i][0] = i
	}
	for j := 0; j <= la; j++ {
		matrix[0][j] = j
	}
	for i := 1; i <= lb; i++ {
		for j := 1; j <= la; j++ {
			if b[i-1] == a[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				min := matrix[i-1][j-1] + 1
				if matrix[i][j-1]+1 < min {
					min = matrix[i][j-1] + 1
				}
				if matrix[i-1][j]+1 < min {
					min = matrix[i-1][j] + 1
				}
				matrix[i][j] = min
			}
		}
	}
	return matrix[lb][la]
}
