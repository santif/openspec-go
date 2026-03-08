package cli

import (
	"fmt"
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
	viewCmd := &cobra.Command{
		Use:   "view",
		Short: "Display an interactive dashboard of specs and changes",
		RunE:  runView,
	}
	rootCmd.AddCommand(viewCmd)
}

func runView(cmd *cobra.Command, args []string) error {
	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, coreconfig.OpenSpecDirName)
	if !utils.DirectoryExists(openspecDir) {
		return fmt.Errorf("no openspec directory found. Run 'openspec init' first")
	}

	changes := utils.GetActiveChangeIDs(projectRoot)
	specs := utils.GetSpecIDs(projectRoot)

	header := color.New(color.FgCyan, color.Bold)
	header.Println("\n OpenSpec Dashboard")
	header.Println(strings.Repeat("-", 50))

	// Specs section
	fmt.Printf("\n  %s Specs (%d)\n", color.CyanString("[S]"), len(specs))
	for _, id := range specs {
		fmt.Printf("    %s %s\n", color.GreenString("*"), id)
	}
	if len(specs) == 0 {
		fmt.Println("    (none)")
	}

	// Changes section
	fmt.Printf("\n  %s Active Changes (%d)\n", color.CyanString("[C]"), len(changes))

	// Resolve schema for progress tracking
	schemaName := "spec-driven"
	cfg := projectconfig.ReadProjectConfig(projectRoot)
	if cfg != nil && cfg.Schema != "" {
		schemaName = cfg.Schema
	}

	for _, id := range changes {
		changeDir := filepath.Join(openspecDir, "changes", id)

		schema, err := artifactgraph.ResolveSchema(schemaName, projectRoot)
		if err != nil {
			fmt.Printf("    %s %s\n", color.YellowString("*"), id)
			continue
		}

		graph := artifactgraph.NewGraphFromSchema(schema)
		completed := artifactgraph.DetectCompleted(graph, changeDir)
		totalArtifacts := len(graph.GetAllArtifacts())
		completedCount := len(completed)

		// Progress bar
		bar := progressBar(completedCount, totalArtifacts, 20)

		if graph.IsComplete(completed) {
			fmt.Printf("    %s %s %s (%d/%d)\n", color.GreenString("OK"), id, bar, completedCount, totalArtifacts)
		} else {
			fmt.Printf("    %s %s %s (%d/%d)\n", color.YellowString(".."), id, bar, completedCount, totalArtifacts)
		}

		// Task progress if tasks.md exists
		tasksPath := filepath.Join(changeDir, "tasks.md")
		if utils.FileExists(tasksPath) {
			content, err := utils.ReadFile(tasksPath)
			if err == nil {
				tp := utils.CountTasks(content)
				if tp.Total > 0 {
					taskBar := progressBar(tp.Completed, tp.Total, 15)
					fmt.Printf("      Tasks: %s (%d/%d)\n", taskBar, tp.Completed, tp.Total)
				}
			}
		}
	}

	if len(changes) == 0 {
		fmt.Println("    (none)")
	}

	// Archived changes
	archived := utils.GetArchivedChangeIDs(projectRoot)
	if len(archived) > 0 {
		fmt.Printf("\n  %s Archived (%d)\n", color.CyanString("[A]"), len(archived))
		for _, id := range archived {
			dim := color.New(color.FgHiBlack)
			fmt.Printf("    %s %s\n", dim.Sprint("-"), dim.Sprint(id))
		}
	}

	fmt.Println()
	return nil
}

func progressBar(completed, total, width int) string {
	if total == 0 {
		return strings.Repeat(".", width)
	}
	filled := (completed * width) / total
	if filled > width {
		filled = width
	}
	return color.GreenString(strings.Repeat("#", filled)) + strings.Repeat(".", width-filled)
}
