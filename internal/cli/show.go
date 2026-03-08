package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/parsers"
	"github.com/santif/openspec-go/internal/utils"
)

func init() {
	showCmd := &cobra.Command{
		Use:   "show [item-name]",
		Short: "Show a change or spec",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runShow,
	}
	showCmd.Flags().Bool("json", false, "Output as JSON")
	showCmd.Flags().String("type", "", "Specify item type when ambiguous: change|spec")
	showCmd.Flags().Bool("deltas-only", false, "Show only deltas (JSON only, change)")
	showCmd.Flags().Bool("requirements", false, "JSON only: Show only requirements (exclude scenarios)")
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	itemType, _ := cmd.Flags().GetString("type")
	deltasOnly, _ := cmd.Flags().GetBool("deltas-only")

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, config.OpenSpecDirName)

	var itemName string
	if len(args) > 0 {
		itemName = args[0]
	}

	if itemName == "" {
		// Try to select interactively or list available
		changes := utils.GetActiveChangeIDs(projectRoot)
		specs := utils.GetSpecIDs(projectRoot)
		if len(changes) == 0 && len(specs) == 0 {
			return fmt.Errorf("no changes or specs found")
		}
		// Default to first change or spec
		if len(changes) > 0 {
			itemName = changes[0]
			itemType = "change"
		} else {
			itemName = specs[0]
			itemType = "spec"
		}
	}

	// Auto-detect type if not specified
	if itemType == "" {
		changePath := filepath.Join(openspecDir, "changes", itemName, "proposal.md")
		specPath := filepath.Join(openspecDir, "specs", itemName, "spec.md")

		changeExists := utils.FileExists(changePath)
		specExists := utils.FileExists(specPath)

		if changeExists && specExists {
			itemType = "change" // Prefer change when ambiguous
		} else if changeExists {
			itemType = "change"
		} else if specExists {
			itemType = "spec"
		} else {
			return fmt.Errorf("item %q not found as change or spec", itemName)
		}
	}

	if itemType == "change" {
		return showChange(openspecDir, itemName, jsonOutput, deltasOnly)
	}
	return showSpec(openspecDir, itemName, jsonOutput)
}

func showChange(openspecDir, name string, jsonOutput, deltasOnly bool) error {
	proposalPath := filepath.Join(openspecDir, "changes", name, "proposal.md")
	content, err := utils.ReadFile(proposalPath)
	if err != nil {
		return fmt.Errorf("cannot read change %q: %w", name, err)
	}

	changeDir := filepath.Join(openspecDir, "changes", name)
	change, err := parsers.ParseChangeWithDeltas(name, content, changeDir)
	if err != nil {
		return fmt.Errorf("error parsing change %q: %w", name, err)
	}

	if jsonOutput {
		if deltasOnly {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(change.Deltas)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(change)
	}

	// Print markdown
	fmt.Println(content)
	return nil
}

func showSpec(openspecDir, name string, jsonOutput bool) error {
	specPath := filepath.Join(openspecDir, "specs", name, "spec.md")
	content, err := utils.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("cannot read spec %q: %w", name, err)
	}

	spec, err := parsers.ParseSpec(name, content)
	if err != nil {
		return fmt.Errorf("error parsing spec %q: %w", name, err)
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(spec)
	}

	// Print markdown
	fmt.Println(content)
	return nil
}
