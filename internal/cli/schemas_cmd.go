package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/artifactgraph"
)

func init() {
	schemasCmd := &cobra.Command{
		Use:   "schemas",
		Short: "List available workflow schemas with descriptions",
		RunE:  runSchemas,
	}
	schemasCmd.Flags().Bool("json", false, "Output as JSON (for agent use)")
	rootCmd.AddCommand(schemasCmd)
}

type schemaInfo struct {
	Name        string `json:"name"`
	IsBuiltIn   bool   `json:"isBuiltIn"`
	Description string `json:"description,omitempty"`
}

func runSchemas(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot := "."
	sources := artifactgraph.ListAvailableSchemas(projectRoot)

	var infos []schemaInfo
	for _, s := range sources {
		info := schemaInfo{
			Name:      s.Name,
			IsBuiltIn: s.IsBuiltIn,
		}

		// Try to load description
		schema, err := artifactgraph.ResolveSchema(s.Name, projectRoot)
		if err == nil {
			info.Description = schema.Description
		}

		infos = append(infos, info)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(infos)
	}

	fmt.Println("Available schemas:")
	fmt.Println()
	for _, info := range infos {
		source := "built-in"
		if !info.IsBuiltIn {
			source = "project-local"
		}
		fmt.Printf("  %s (%s)", info.Name, source)
		if info.Description != "" {
			fmt.Printf(" - %s", info.Description)
		}
		fmt.Println()
	}
	fmt.Println()

	return nil
}
