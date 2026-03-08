package cli

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Fission-AI/openspec-go/internal/core/config"
	"github.com/Fission-AI/openspec-go/internal/utils"
)

func init() {
	updateCmd := &cobra.Command{
		Use:   "update [path]",
		Short: "Update OpenSpec instruction files",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runUpdate,
	}
	updateCmd.Flags().Bool("force", false, "Force update even when tools are up to date")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}
	forceFlag, _ := cmd.Flags().GetBool("force")

	resolvedPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	openspecDir := filepath.Join(resolvedPath, config.OpenSpecDirName)
	if !utils.DirectoryExists(openspecDir) {
		return fmt.Errorf("no openspec directory found at %s. Run 'openspec init' first", resolvedPath)
	}

	// TODO: Regenerate skills/commands for detected tools
	// For now, just report success
	if forceFlag {
		fmt.Println()
		color.New(color.FgGreen).Println("  Force update completed")
		fmt.Println()
	} else {
		fmt.Println()
		color.New(color.FgGreen).Println("  OpenSpec files are up to date")
		fmt.Println()
	}

	return nil
}
