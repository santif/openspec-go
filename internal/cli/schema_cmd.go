package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	coreconfig "github.com/Fission-AI/openspec-go/internal/core/config"
	"github.com/Fission-AI/openspec-go/internal/core/artifactgraph"
	"github.com/Fission-AI/openspec-go/internal/core/projectconfig"
	"github.com/Fission-AI/openspec-go/internal/utils"
)

func init() {
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage workflow schemas",
	}

	schemaCmd.AddCommand(&cobra.Command{
		Use:   "which",
		Short: "Show which schema is currently active",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			cfg := projectconfig.ReadProjectConfig(projectRoot)
			schemaName := "spec-driven"
			if cfg != nil && cfg.Schema != "" {
				schemaName = cfg.Schema
			}
			fmt.Println(schemaName)
			return nil
		},
	})

	schemaValidateCmd := &cobra.Command{
		Use:   "validate [schema-name]",
		Short: "Validate a schema file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			schemaName := "spec-driven"
			if len(args) > 0 {
				schemaName = args[0]
			}

			schema, err := artifactgraph.ResolveSchema(schemaName, projectRoot)
			if err != nil {
				color.New(color.FgRed).Fprintf(os.Stderr, "Schema %q is invalid: %v\n", schemaName, err)
				os.Exit(1)
			}

			color.New(color.FgGreen).Printf("Schema %q is valid\n", schema.Name)
			fmt.Printf("  Version: %d\n", schema.Version)
			fmt.Printf("  Artifacts: %d\n", len(schema.Artifacts))
			return nil
		},
	}
	schemaCmd.AddCommand(schemaValidateCmd)

	schemaForkCmd := &cobra.Command{
		Use:   "fork [schema-name]",
		Short: "Copy a built-in schema to project-local for customization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectRoot := "."
			schemaName := "spec-driven"
			if len(args) > 0 {
				schemaName = args[0]
			}

			// Read built-in schema
			schema, err := artifactgraph.ResolveSchema(schemaName, projectRoot)
			if err != nil {
				return fmt.Errorf("schema %q not found: %w", schemaName, err)
			}

			// Write to project-local
			destDir := fmt.Sprintf("%s/%s/schemas/%s", projectRoot, coreconfig.OpenSpecDirName, schemaName)
			if utils.DirectoryExists(destDir) {
				return fmt.Errorf("project-local schema %q already exists at %s", schemaName, destDir)
			}

			if err := utils.EnsureDir(destDir + "/templates"); err != nil {
				return err
			}

			// Write schema.yaml - try to read the raw content from embedded
			content := fmt.Sprintf("name: %s\nversion: %d\n", schema.Name, schema.Version)
			if err := utils.WriteFile(destDir+"/schema.yaml", content); err != nil {
				return err
			}

			// Copy templates
			for _, a := range schema.Artifacts {
				templateContent, err := artifactgraph.ReadTemplate(schemaName, a.Template)
				if err != nil {
					continue
				}
				if err := utils.WriteFile(destDir+"/templates/"+a.Template, templateContent); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to copy template %s: %v\n", a.Template, err)
				}
			}

			color.New(color.FgGreen).Printf("Forked schema %q to %s\n", schemaName, destDir)
			fmt.Println("  You can now customize it for your project.")
			return nil
		},
	}
	schemaCmd.AddCommand(schemaForkCmd)

	rootCmd.AddCommand(schemaCmd)
}
