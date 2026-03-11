package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/artifactgraph"
	coreconfig "github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/utils"
)

func init() {
	instructionsCmd := &cobra.Command{
		Use:   "instructions [artifact]",
		Short: "Output enriched instructions for creating an artifact or applying tasks",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInstructions,
	}
	instructionsCmd.Flags().String("change", "", "Change name")
	instructionsCmd.Flags().String("schema", "", "Schema override")
	instructionsCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(instructionsCmd)
}

func runInstructions(cmd *cobra.Command, args []string) error {
	changeName, _ := cmd.Flags().GetString("change")
	schemaFlag, _ := cmd.Flags().GetString("schema")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot := "."

	// Read project config once for schema, context, rules, and conditionals
	cfg := projectconfig.ReadProjectConfig(projectRoot)

	// Resolve schema
	schemaName := schemaFlag
	if schemaName == "" {
		if changeName != "" {
			changeDir := filepath.Join(projectRoot, coreconfig.OpenSpecDirName, "changes", changeName)
			meta, err := utils.ReadChangeMetadata(changeDir)
			if err == nil && meta.Schema != "" {
				schemaName = meta.Schema
			}
		}
	}
	if schemaName == "" {
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

	// Get project context, rules, and conditional keywords
	var context string
	var rules map[string][]string
	var conditionals *projectconfig.ConditionalsConfig
	if cfg != nil {
		context = cfg.Context
		rules = cfg.Rules
		if cfg.Keywords != nil && cfg.Keywords.Conditionals != nil {
			resolved := projectconfig.ResolveConditionals(cfg.Keywords)
			conditionals = &resolved
		}
	}

	// Special case: "apply"
	var artifactID string
	if len(args) > 0 {
		artifactID = args[0]
	}

	if artifactID == "apply" {
		var applyInstruction *artifactgraph.EnrichedInstruction
		applyInstruction, err = artifactgraph.LoadApplyInstruction(graph, context, rules, conditionals)
		if err != nil {
			return err
		}
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(applyInstruction)
		}
		fmt.Println(applyInstruction.Instruction)
		return nil
	}

	if artifactID == "" {
		// List available artifacts
		fmt.Println("Available artifacts:")
		for _, a := range graph.GetAllArtifacts() {
			fmt.Printf("  %s - %s\n", a.ID, a.Description)
		}
		fmt.Println("\nSpecial:")
		fmt.Println("  apply - Get apply/implementation instructions")
		return nil
	}

	instruction, err := artifactgraph.LoadEnrichedInstruction(graph, artifactID, context, rules, conditionals)
	if err != nil {
		return err
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(instruction)
	}

	fmt.Println(instruction.Instruction)
	return nil
}
