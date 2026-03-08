package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/utils"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List items (changes by default). Use --specs to list specs.",
		RunE:  runList,
	}
	listCmd.Flags().Bool("specs", false, "List specs instead of changes")
	listCmd.Flags().Bool("changes", false, "List changes explicitly (default)")
	listCmd.Flags().String("sort", "recent", `Sort order: "recent" (default) or "name"`)
	listCmd.Flags().Bool("json", false, "Output as JSON (for programmatic use)")
	rootCmd.AddCommand(listCmd)
}

type listItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Path    string `json:"path"`
	ModTime string `json:"modTime,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	specs, _ := cmd.Flags().GetBool("specs")
	sortOrder, _ := cmd.Flags().GetString("sort")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	mode := "changes"
	if specs {
		mode = "specs"
	}

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, config.OpenSpecDirName)
	if !utils.DirectoryExists(openspecDir) {
		return fmt.Errorf("no openspec directory found. Run 'openspec init' first")
	}

	var items []listItem

	if mode == "changes" {
		ids := utils.GetActiveChangeIDs(projectRoot)
		for _, id := range ids {
			changePath := filepath.Join(openspecDir, "changes", id)
			modTime := getModTime(changePath)
			items = append(items, listItem{
				ID:      id,
				Type:    "change",
				Path:    changePath,
				ModTime: modTime,
			})
		}
	} else {
		ids := utils.GetSpecIDs(projectRoot)
		for _, id := range ids {
			specPath := filepath.Join(openspecDir, "specs", id)
			modTime := getModTime(specPath)
			items = append(items, listItem{
				ID:      id,
				Type:    "spec",
				Path:    specPath,
				ModTime: modTime,
			})
		}
	}

	// Sort
	if sortOrder == "name" {
		sort.Slice(items, func(i, j int) bool {
			return items[i].ID < items[j].ID
		})
	} else {
		// recent: sort by mod time descending
		sort.Slice(items, func(i, j int) bool {
			return items[i].ModTime > items[j].ModTime
		})
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}

	if len(items) == 0 {
		fmt.Printf("No %s found.\n", mode)
		return nil
	}

	header := color.New(color.FgCyan, color.Bold)
	header.Printf("\n %s (%d)\n\n", titleCase(mode), len(items))

	for _, item := range items {
		fmt.Printf("  %s %s\n", color.GreenString("*"), item.ID)
	}
	fmt.Println()

	return nil
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func getModTime(dirPath string) string {
	info, err := os.Stat(dirPath)
	if err != nil {
		return ""
	}
	return info.ModTime().Format(time.RFC3339)
}
