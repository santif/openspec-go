package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/Fission-AI/openspec-go/internal/core/config"
	"github.com/Fission-AI/openspec-go/internal/utils"
)

func init() {
	initCmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize OpenSpec in your project",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInit,
	}

	availableToolIDs := []string{}
	for _, tool := range config.AITools {
		if tool.SkillsDir != "" {
			availableToolIDs = append(availableToolIDs, tool.Value)
		}
	}
	toolsDesc := fmt.Sprintf("Configure AI tools non-interactively. Use \"all\", \"none\", or a comma-separated list of: %s", strings.Join(availableToolIDs, ", "))

	initCmd.Flags().String("tools", "", toolsDesc)
	initCmd.Flags().Bool("force", false, "Auto-cleanup legacy files without prompting")
	initCmd.Flags().String("profile", "", "Override global config profile (core or custom)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}
	toolsFlag, _ := cmd.Flags().GetString("tools")
	profileFlag, _ := cmd.Flags().GetString("profile")

	resolvedPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Create openspec directory structure
	openspecDir := filepath.Join(resolvedPath, config.OpenSpecDirName)
	dirs := []string{
		openspecDir,
		filepath.Join(openspecDir, "specs"),
		filepath.Join(openspecDir, "changes"),
	}

	for _, dir := range dirs {
		if err := utils.EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write config.yaml
	configPath := filepath.Join(openspecDir, "config.yaml")
	if !utils.FileExists(configPath) {
		schemaName := "spec-driven"
		configContent := fmt.Sprintf("schema: %s\n", schemaName)
		if err := utils.WriteFile(configPath, configContent); err != nil {
			return fmt.Errorf("failed to write config.yaml: %w", err)
		}
	}

	// Resolve selected tools
	selectedTools := resolveTools(toolsFlag)

	// Generate tool-specific files (skills/commands)
	for _, tool := range selectedTools {
		if tool.SkillsDir == "" {
			continue
		}
		skillsDir := filepath.Join(resolvedPath, tool.SkillsDir)
		if err := utils.EnsureDir(skillsDir); err != nil {
			continue
		}
	}

	// Print welcome
	fmt.Println()
	header := color.New(color.FgCyan, color.Bold)
	header.Println("  OpenSpec initialized!")
	fmt.Println()

	success := color.New(color.FgGreen)
	success.Printf("  Created %s/\n", config.OpenSpecDirName)
	success.Printf("  Created %s/config.yaml\n", config.OpenSpecDirName)
	if len(selectedTools) > 0 {
		toolNames := make([]string, 0, len(selectedTools))
		for _, t := range selectedTools {
			toolNames = append(toolNames, t.SuccessLabel)
		}
		success.Printf("  Configured for: %s\n", strings.Join(toolNames, ", "))
	}
	fmt.Println()

	// Profile info
	if profileFlag != "" {
		fmt.Printf("  Profile: %s\n", profileFlag)
	}

	fmt.Printf("  Next: Create a change with 'openspec new change <name>'\n\n")

	return nil
}

func resolveTools(toolsFlag string) []config.AIToolOption {
	if toolsFlag == "" {
		return nil
	}
	if toolsFlag == "none" {
		return nil
	}

	var selected []config.AIToolOption
	if toolsFlag == "all" {
		for _, tool := range config.AITools {
			if tool.Available && tool.SkillsDir != "" {
				selected = append(selected, tool)
			}
		}
		return selected
	}

	ids := strings.Split(toolsFlag, ",")
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[strings.TrimSpace(id)] = true
	}

	for _, tool := range config.AITools {
		if idSet[tool.Value] && tool.Available {
			selected = append(selected, tool)
		}
	}

	return selected
}
