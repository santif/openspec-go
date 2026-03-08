package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	feedbackCmd := &cobra.Command{
		Use:   "feedback <message>",
		Short: "Submit feedback about OpenSpec",
		Args:  cobra.ExactArgs(1),
		RunE:  runFeedback,
	}
	feedbackCmd.Flags().String("body", "", "Detailed description for the feedback")
	rootCmd.AddCommand(feedbackCmd)
}

func runFeedback(cmd *cobra.Command, args []string) error {
	message := args[0]
	body, _ := cmd.Flags().GetString("body")

	// Check if gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is required for feedback. Install from https://cli.github.com")
	}

	ghArgs := []string{
		"issue", "create",
		"--repo", "Fission-AI/OpenSpec",
		"--title", message,
	}

	if body != "" {
		ghArgs = append(ghArgs, "--body", body)
	} else {
		ghArgs = append(ghArgs, "--body", fmt.Sprintf("Feedback: %s", message))
	}

	ghArgs = append(ghArgs, "--label", "feedback")

	fmt.Printf("Creating issue: %s\n", message)
	c := exec.Command("gh", ghArgs...)
	output, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create issue: %s\n%s", err, strings.TrimSpace(string(output)))
	}

	fmt.Printf("Done: %s\n", strings.TrimSpace(string(output)))
	return nil
}
