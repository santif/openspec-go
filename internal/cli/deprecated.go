package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// Deprecated change subcommands
	changeCmd := &cobra.Command{
		Use:   "change",
		Short: "Manage OpenSpec change proposals",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stderr, "Warning: The \"openspec change ...\" commands are deprecated. Prefer verb-first commands (e.g., \"openspec list\", \"openspec validate --changes\").")
		},
	}

	changeShowCmd := &cobra.Command{
		Use:   "show [change-name]",
		Short: "Show a change proposal (DEPRECATED: use \"openspec show\")",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec show\" instead.")
			return runShow(cmd, args)
		},
	}
	changeShowCmd.Flags().Bool("json", false, "Output as JSON")
	changeShowCmd.Flags().String("type", "change", "Item type")
	changeShowCmd.Flags().Bool("deltas-only", false, "Show only deltas (JSON only)")
	changeShowCmd.Flags().Bool("requirements", false, "JSON only: Show only requirements")
	changeCmd.AddCommand(changeShowCmd)

	changeListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all active changes (DEPRECATED: use \"openspec list\")",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec list\" instead.")
			return runList(cmd, args)
		},
	}
	changeListCmd.Flags().Bool("specs", false, "List specs instead of changes")
	changeListCmd.Flags().Bool("changes", false, "List changes explicitly (default)")
	changeListCmd.Flags().String("sort", "recent", "Sort order")
	changeListCmd.Flags().Bool("json", false, "Output as JSON")
	changeCmd.AddCommand(changeListCmd)

	changeValidateCmd := &cobra.Command{
		Use:   "validate [change-name]",
		Short: "Validate a change proposal (DEPRECATED: use \"openspec validate\")",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec validate\" instead.")
			return runValidate(cmd, args)
		},
	}
	changeValidateCmd.Flags().Bool("all", false, "Validate all")
	changeValidateCmd.Flags().Bool("changes", true, "Validate changes")
	changeValidateCmd.Flags().Bool("specs", false, "Validate specs")
	changeValidateCmd.Flags().String("type", "change", "Item type")
	changeValidateCmd.Flags().Bool("strict", false, "Strict mode")
	changeValidateCmd.Flags().Bool("json", false, "Output as JSON")
	changeCmd.AddCommand(changeValidateCmd)

	rootCmd.AddCommand(changeCmd)

	// Deprecated spec subcommands
	specCmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage OpenSpec specifications",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stderr, "Warning: The \"openspec spec ...\" commands are deprecated. Prefer verb-first commands.")
		},
	}

	specShowCmd := &cobra.Command{
		Use:   "show [spec-name]",
		Short: "Show a specification (DEPRECATED: use \"openspec show --type spec\")",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec show --type spec\" instead.")
			return runShow(cmd, args)
		},
	}
	specShowCmd.Flags().Bool("json", false, "Output as JSON")
	specShowCmd.Flags().String("type", "spec", "Item type")
	specShowCmd.Flags().Bool("deltas-only", false, "Show only deltas (JSON only)")
	specShowCmd.Flags().Bool("requirements", false, "JSON only: Show only requirements")
	specCmd.AddCommand(specShowCmd)

	specListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all specs (DEPRECATED: use \"openspec list --specs\")",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec list --specs\" instead.")
			return runList(cmd, args)
		},
	}
	specListCmd.Flags().Bool("specs", true, "List specs")
	specListCmd.Flags().Bool("changes", false, "List changes")
	specListCmd.Flags().String("sort", "recent", "Sort order")
	specListCmd.Flags().Bool("json", false, "Output as JSON")
	specCmd.AddCommand(specListCmd)

	specValidateCmd := &cobra.Command{
		Use:   "validate [spec-name]",
		Short: "Validate a specification (DEPRECATED: use \"openspec validate\")",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "Use \"openspec validate\" instead.")
			return runValidate(cmd, args)
		},
	}
	specValidateCmd.Flags().Bool("all", false, "Validate all")
	specValidateCmd.Flags().Bool("changes", false, "Validate changes")
	specValidateCmd.Flags().Bool("specs", true, "Validate specs")
	specValidateCmd.Flags().String("type", "spec", "Item type")
	specValidateCmd.Flags().Bool("strict", false, "Strict mode")
	specValidateCmd.Flags().Bool("json", false, "Output as JSON")
	specCmd.AddCommand(specValidateCmd)

	rootCmd.AddCommand(specCmd)
}
