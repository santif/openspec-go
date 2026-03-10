package cli

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/commandgen"
	_ "github.com/santif/openspec-go/internal/core/commandgen/adapters"
	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/globalconfig"
	"github.com/santif/openspec-go/internal/core/profiles"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/utils"
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

	gcfg := globalconfig.GetGlobalConfig()
	workflows := profiles.GetProfileWorkflows(gcfg.Profile, gcfg.Workflows)

	// Read project config for conditional keywords
	var conditionals *projectconfig.ConditionalsConfig
	if cfg := projectconfig.ReadProjectConfig(resolvedPath); cfg != nil && cfg.Keywords != nil {
		conditionals = cfg.Keywords.Conditionals
	}

	// Detect which tools have skill directories present
	var updated int
	for _, tool := range config.AITools {
		if tool.SkillsDir == "" {
			continue
		}
		// Check if the tool's skills directory exists in the project
		toolDir := filepath.Join(resolvedPath, tool.SkillsDir)
		if !utils.DirectoryExists(toolDir) && !forceFlag {
			continue
		}
		if !utils.DirectoryExists(toolDir) {
			continue
		}

		files, err := commandgen.GenerateForTool(tool.Value, workflows, gcfg.Delivery, version, conditionals)
		if err != nil {
			continue
		}
		for _, f := range files {
			dir := f.Dir
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(resolvedPath, dir)
			}
			if err := utils.EnsureDir(dir); err != nil {
				continue
			}
			_ = utils.WriteFile(filepath.Join(dir, f.FileName), f.Content)
		}
		updated++
	}

	fmt.Println()
	if updated > 0 {
		color.New(color.FgGreen).Printf("  Updated skills for %d tool(s)\n", updated)
	} else {
		color.New(color.FgGreen).Println("  OpenSpec files are up to date")
	}
	fmt.Println()

	return nil
}
