package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	coreconfig "github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/specsapply"
	"github.com/santif/openspec-go/internal/core/validation"
	"github.com/santif/openspec-go/internal/utils"
)

func init() {
	archiveCmd := &cobra.Command{
		Use:   "archive [change-name]",
		Short: "Archive a completed change and update main specs",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runArchive,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return utils.GetActiveChangeIDs("."), cobra.ShellCompDirectiveNoFileComp
		},
	}
	archiveCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	archiveCmd.Flags().Bool("skip-specs", false, "Skip spec update operations")
	archiveCmd.Flags().Bool("no-validate", false, "Skip validation (not recommended)")
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	skipSpecs, _ := cmd.Flags().GetBool("skip-specs")
	noValidate, _ := cmd.Flags().GetBool("no-validate")

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, coreconfig.OpenSpecDirName)

	var changeName string
	if len(args) > 0 {
		changeName = args[0]
	} else {
		changes := utils.GetActiveChangeIDs(projectRoot)
		if len(changes) == 0 {
			return fmt.Errorf("no active changes found")
		}
		if len(changes) == 1 {
			changeName = changes[0]
		} else {
			return fmt.Errorf("multiple changes found, specify one: %v", changes)
		}
	}

	changeDir := filepath.Join(openspecDir, "changes", changeName)
	if !utils.DirectoryExists(changeDir) {
		return fmt.Errorf("change %q not found", changeName)
	}

	// Validate unless skipped
	if !noValidate {
		v := validation.NewValidator(false)
		proposalPath := filepath.Join(changeDir, "proposal.md")
		if utils.FileExists(proposalPath) {
			report := v.ValidateChange(proposalPath)
			if !report.Valid {
				fmt.Println()
				color.New(color.FgRed).Printf("  Validation failed for %s\n", changeName)
				for _, issue := range report.Issues {
					if issue.Level == validation.LevelError {
						fmt.Printf("    ERROR [%s] %s\n", issue.Path, issue.Message)
					}
				}
				fmt.Println()
				if !yes {
					return fmt.Errorf("fix validation errors before archiving, or use --no-validate")
				}
			}
		}
	}

	// Apply delta specs to main specs
	if !skipSpecs {
		result, err := specsapply.ApplySpecs(projectRoot, changeName, specsapply.ApplyOptions{
			SkipValidation: noValidate,
		})
		if err != nil {
			fmt.Println()
			color.New(color.FgRed).Printf("  Spec application failed: %s\n", err)
			fmt.Println("  Aborted. No files were changed.")
			fmt.Println()
			return fmt.Errorf("spec application failed: %w", err)
		}
		if !result.NoChanges {
			fmt.Println()
			total := result.Totals.Added + result.Totals.Modified + result.Totals.Removed + result.Totals.Renamed
			color.New(color.FgGreen).Printf("  Applied %d spec update(s)\n", total)
		}
	}

	// Move change to archive
	archiveDir := filepath.Join(openspecDir, "archive")
	if err := utils.EnsureDir(archiveDir); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	archiveDest := filepath.Join(archiveDir, date+"-"+changeName)
	if utils.DirectoryExists(archiveDest) {
		return fmt.Errorf("archive already contains %q", changeName)
	}

	if err := os.Rename(changeDir, archiveDest); err != nil {
		return fmt.Errorf("failed to move change to archive: %w", err)
	}

	fmt.Println()
	color.New(color.FgGreen).Printf("  Archived %s\n", changeName)
	fmt.Println()

	return nil
}
