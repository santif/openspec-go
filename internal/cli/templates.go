package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Fission-AI/openspec-go/internal/core/artifactgraph"
	"github.com/Fission-AI/openspec-go/internal/core/projectconfig"
)

func init() {
	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: "Show resolved template paths for all artifacts in a schema",
		RunE:  runTemplates,
	}
	templatesCmd.Flags().String("schema", "", "Schema to use (default: spec-driven)")
	templatesCmd.Flags().Bool("json", false, "Output as JSON mapping artifact IDs to template paths")
	rootCmd.AddCommand(templatesCmd)
}

func runTemplates(cmd *cobra.Command, args []string) error {
	schemaFlag, _ := cmd.Flags().GetString("schema")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot := "."

	schemaName := schemaFlag
	if schemaName == "" {
		cfg := projectconfig.ReadProjectConfig(projectRoot)
		if cfg != nil && cfg.Schema != "" {
			schemaName = cfg.Schema
		}
	}
	if schemaName == "" {
		schemaName = "spec-driven"
	}

	schema, err := artifactgraph.ResolveSchema(schemaName, projectRoot)
	if err != nil {
		return fmt.Errorf("schema %q not found: %w", schemaName, err)
	}

	graph := artifactgraph.NewGraphFromSchema(schema)

	templateMap := make(map[string]string)
	for _, a := range graph.GetAllArtifacts() {
		path := artifactgraph.ResolveTemplatePath(schemaName, a.Template, projectRoot)
		templateMap[a.ID] = path
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(templateMap)
	}

	fmt.Printf("Templates for schema %q:\n\n", schemaName)
	for _, a := range graph.GetAllArtifacts() {
		fmt.Printf("  %s -> %s\n", a.ID, templateMap[a.ID])
	}
	fmt.Println()

	return nil
}
