package artifactgraph

import (
	"fmt"
	"os"
	"path/filepath"

	builtinSchemas "github.com/santif/openspec-go/schemas"
	"github.com/santif/openspec-go/internal/core/globalconfig"
)

// SchemaSource describes where a schema was found.
type SchemaSource struct {
	Name      string
	IsBuiltIn bool
	Path      string // empty for built-in (use embed)
}

// ResolveSchema resolves a schema by name, checking user override, project-local, then built-in.
func ResolveSchema(schemaName, projectRoot string) (*SchemaYaml, error) {
	// 1. Try user override (~/.local/share/openspec/schemas/<name>/schema.yaml)
	userDataDir := globalconfig.GetGlobalDataDir()
	userSchemaPath := filepath.Join(userDataDir, "schemas", schemaName, "schema.yaml")
	if _, err := os.Stat(userSchemaPath); err == nil {
		return LoadSchema(userSchemaPath)
	}

	// 2. Try project-local (openspec/schemas/<name>/schema.yaml)
	projectSchemaPath := filepath.Join(projectRoot, "openspec", "schemas", schemaName, "schema.yaml")
	if _, err := os.Stat(projectSchemaPath); err == nil {
		return LoadSchema(projectSchemaPath)
	}

	// 3. Try built-in (embedded)
	embeddedPath := schemaName + "/schema.yaml"
	data, err := builtinSchemas.BuiltinSchemas.ReadFile(embeddedPath)
	if err == nil {
		return ParseSchema(string(data))
	}

	return nil, fmt.Errorf("schema %q not found", schemaName)
}

// ListAvailableSchemas returns all schemas from built-in and project-local sources.
func ListAvailableSchemas(projectRoot string) []SchemaSource {
	var sources []SchemaSource

	// Built-in schemas
	entries, err := builtinSchemas.BuiltinSchemas.ReadDir(".")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				sources = append(sources, SchemaSource{
					Name:      entry.Name(),
					IsBuiltIn: true,
				})
			}
		}
	}

	// Project-local schemas
	projectSchemasDir := filepath.Join(projectRoot, "openspec", "schemas")
	if entries, err := os.ReadDir(projectSchemasDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Check if schema.yaml exists
				schemaPath := filepath.Join(projectSchemasDir, entry.Name(), "schema.yaml")
				if _, err := os.Stat(schemaPath); err == nil {
					sources = append(sources, SchemaSource{
						Name:      entry.Name(),
						IsBuiltIn: false,
						Path:      schemaPath,
					})
				}
			}
		}
	}

	return sources
}

// ResolveTemplatePath returns the filesystem path for a template, checking project-local then user override.
func ResolveTemplatePath(schemaName, artifactTemplate, projectRoot string) string {
	// 1. Project-local template
	projectTemplatePath := filepath.Join(projectRoot, "openspec", "schemas", schemaName, "templates", artifactTemplate)
	if _, err := os.Stat(projectTemplatePath); err == nil {
		return projectTemplatePath
	}

	// 2. User override template
	userDataDir := globalconfig.GetGlobalDataDir()
	userTemplatePath := filepath.Join(userDataDir, "schemas", schemaName, "templates", artifactTemplate)
	if _, err := os.Stat(userTemplatePath); err == nil {
		return userTemplatePath
	}

	// 3. Built-in template (return the embedded path marker)
	return fmt.Sprintf("builtin:%s/templates/%s", schemaName, artifactTemplate)
}

// ReadTemplate reads a template from built-in embedded schemas.
func ReadTemplate(schemaName, templateName string) (string, error) {
	// Try built-in
	embeddedPath := schemaName + "/templates/" + templateName
	data, err := builtinSchemas.BuiltinSchemas.ReadFile(embeddedPath)
	if err == nil {
		return string(data), nil
	}
	return "", fmt.Errorf("template %q not found in schema %q", templateName, schemaName)
}
