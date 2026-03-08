package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	coreconfig "github.com/Fission-AI/openspec-go/internal/core/config"
	"github.com/Fission-AI/openspec-go/internal/core/artifactgraph"
	"github.com/Fission-AI/openspec-go/internal/core/projectconfig"
	"github.com/Fission-AI/openspec-go/internal/utils"
)

func init() {
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create new items",
	}

	newChangeCmd := &cobra.Command{
		Use:   "change <name>",
		Short: "Create a new change directory",
		Args:  cobra.ExactArgs(1),
		RunE:  runNewChange,
	}
	newChangeCmd.Flags().String("description", "", "Description to add to README.md")
	newChangeCmd.Flags().String("schema", "", "Workflow schema to use (default: spec-driven)")

	newCmd.AddCommand(newChangeCmd)
	rootCmd.AddCommand(newCmd)
}

func runNewChange(cmd *cobra.Command, args []string) error {
	name := args[0]
	description, _ := cmd.Flags().GetString("description")
	schemaFlag, _ := cmd.Flags().GetString("schema")

	// Validate change name
	if !utils.ValidateChangeName(name) {
		return fmt.Errorf("invalid change name %q: must be kebab-case (e.g., my-feature, add-auth)", name)
	}

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, coreconfig.OpenSpecDirName)
	if !utils.DirectoryExists(openspecDir) {
		return fmt.Errorf("no openspec directory found. Run 'openspec init' first")
	}

	// Check if change already exists
	changeDir := filepath.Join(openspecDir, "changes", name)
	if utils.DirectoryExists(changeDir) {
		return fmt.Errorf("change %q already exists", name)
	}

	// Resolve schema
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

	// Verify schema exists
	schema, err := artifactgraph.ResolveSchema(schemaName, projectRoot)
	if err != nil {
		return fmt.Errorf("schema %q not found: %w", schemaName, err)
	}

	// Create change directory
	if err := utils.EnsureDir(changeDir); err != nil {
		return fmt.Errorf("failed to create change directory: %w", err)
	}

	// Write .openspec.yaml metadata
	meta := &utils.ChangeMetadata{
		Schema:  schemaName,
		Created: time.Now().Format("2006-01-02"),
	}
	if err := utils.WriteChangeMetadata(changeDir, meta); err != nil {
		return fmt.Errorf("failed to write change metadata: %w", err)
	}

	// Write initial template files from schema
	graph := artifactgraph.NewGraphFromSchema(schema)
	buildOrder := graph.GetBuildOrder()

	// Write first artifact template (usually proposal.md)
	if len(buildOrder) > 0 {
		firstArtifact := graph.GetArtifact(buildOrder[0])
		if firstArtifact != nil {
			templateContent, err := artifactgraph.ReadTemplate(schemaName, firstArtifact.Template)
			if err == nil {
				templatePath := filepath.Join(changeDir, firstArtifact.Generates)
				if writeErr := utils.WriteFile(templatePath, templateContent); writeErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to write template %s: %v\n", firstArtifact.Generates, writeErr)
				}
			}
		}
	}

	// Write README if description provided
	if description != "" {
		readmePath := filepath.Join(changeDir, "README.md")
		readmeContent := fmt.Sprintf("# %s\n\n%s\n", name, description)
		if err := utils.WriteFile(readmePath, readmeContent); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write README.md: %v\n", err)
		}
	}

	// Print success
	fmt.Println()
	color.New(color.FgGreen).Printf("  Created change: %s\n", name)
	fmt.Printf("    Schema: %s\n", schemaName)
	fmt.Printf("    Path: %s\n", changeDir)
	fmt.Println()
	fmt.Printf("  Next: Edit %s/proposal.md to describe your change\n\n", changeDir)

	return nil
}
