package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Fission-AI/openspec-go/internal/core/artifactgraph"
	coreconfig "github.com/Fission-AI/openspec-go/internal/core/config"
	"github.com/Fission-AI/openspec-go/internal/core/projectconfig"
	"github.com/Fission-AI/openspec-go/internal/utils"
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display artifact completion status for a change",
		RunE:  runStatus,
	}
	statusCmd.Flags().String("change", "", "Change name to show status for")
	statusCmd.Flags().String("schema", "", "Schema override (auto-detected from config.yaml)")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}

type statusOutput struct {
	Change    string           `json:"change"`
	Schema    string           `json:"schema"`
	Artifacts []artifactStatus `json:"artifacts"`
	Complete  bool             `json:"complete"`
}

type artifactStatus struct {
	ID        string   `json:"id"`
	Status    string   `json:"status"` // "done", "ready", "blocked"
	BlockedBy []string `json:"blockedBy,omitempty"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	changeName, _ := cmd.Flags().GetString("change")
	schemaName, _ := cmd.Flags().GetString("schema")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, coreconfig.OpenSpecDirName)

	// Find change
	if changeName == "" {
		changes := utils.GetActiveChangeIDs(projectRoot)
		if len(changes) == 0 {
			if jsonOutput {
				fmt.Println("[]")
				return nil
			}
			fmt.Println("No active changes found.")
			return nil
		}
		changeName = changes[0]
	}

	changeDir := filepath.Join(openspecDir, "changes", changeName)
	if !utils.DirectoryExists(changeDir) {
		return fmt.Errorf("change %q not found", changeName)
	}

	// Resolve schema
	if schemaName == "" {
		// Try change metadata
		meta, err := utils.ReadChangeMetadata(changeDir)
		if err == nil && meta.Schema != "" {
			schemaName = meta.Schema
		}
	}
	if schemaName == "" {
		// Try project config
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
		return fmt.Errorf("failed to resolve schema %q: %w", schemaName, err)
	}

	graph := artifactgraph.NewGraphFromSchema(schema)
	completed := artifactgraph.DetectCompleted(graph, changeDir)
	blocked := graph.GetBlocked(completed)
	buildOrder := graph.GetBuildOrder()

	var artifacts []artifactStatus
	for _, id := range buildOrder {
		status := "blocked"
		var blockedBy []string

		if completed[id] {
			status = "done"
		} else if deps, isBlocked := blocked[id]; isBlocked {
			blockedBy = deps
		} else {
			status = "ready"
		}

		artifacts = append(artifacts, artifactStatus{
			ID:        id,
			Status:    status,
			BlockedBy: blockedBy,
		})
	}

	output := statusOutput{
		Change:    changeName,
		Schema:    schemaName,
		Artifacts: artifacts,
		Complete:  graph.IsComplete(completed),
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Pretty print
	header := color.New(color.FgCyan, color.Bold)
	header.Printf("\n Status: %s (schema: %s)\n\n", changeName, schemaName)

	for _, a := range artifacts {
		switch a.Status {
		case "done":
			fmt.Printf("  %s %s\n", color.GreenString("OK"), a.ID)
		case "ready":
			fmt.Printf("  %s %s\n", color.YellowString("->"), a.ID)
		case "blocked":
			fmt.Printf("  %s %s (blocked by: %s)\n",
				color.RedString("--"), a.ID,
				strings.Join(a.BlockedBy, ", "))
		}
	}

	if output.Complete {
		fmt.Printf("\n  %s All artifacts complete!\n\n", color.GreenString("OK"))
	} else {
		fmt.Println()
	}

	return nil
}
