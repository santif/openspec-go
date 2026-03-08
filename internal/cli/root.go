package cli

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "openspec",
	Short: "AI-native system for spec-driven development",
	Long:  "OpenSpec — AI-native system for spec-driven development",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.Version = version
}

func Execute() error {
	return rootCmd.Execute()
}
