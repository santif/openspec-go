package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/santif/openspec-go/internal/core/config"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/core/validation"
	"github.com/santif/openspec-go/internal/utils"
)

func init() {
	validateCmd := &cobra.Command{
		Use:   "validate [item-name]",
		Short: "Validate changes and specs",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runValidate,
	}
	validateCmd.Flags().Bool("all", false, "Validate all changes and specs")
	validateCmd.Flags().Bool("changes", false, "Validate all changes")
	validateCmd.Flags().Bool("specs", false, "Validate all specs")
	validateCmd.Flags().String("type", "", "Specify item type when ambiguous: change|spec")
	validateCmd.Flags().Bool("strict", false, "Enable strict validation mode")
	validateCmd.Flags().Bool("json", false, "Output validation results as JSON")
	rootCmd.AddCommand(validateCmd)
}

type validationResult struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Report validation.Report `json:"report"`
}

func runValidate(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")
	changesOnly, _ := cmd.Flags().GetBool("changes")
	specsOnly, _ := cmd.Flags().GetBool("specs")
	itemType, _ := cmd.Flags().GetString("type")
	strict, _ := cmd.Flags().GetBool("strict")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	projectRoot := "."
	openspecDir := filepath.Join(projectRoot, config.OpenSpecDirName)
	if !utils.DirectoryExists(openspecDir) {
		return fmt.Errorf("no openspec directory found. Run 'openspec init' first")
	}

	// Read project config for custom keywords
	var keywords []string
	if cfg := projectconfig.ReadProjectConfig(projectRoot); cfg != nil && cfg.Keywords != nil {
		keywords = cfg.Keywords.Normative
	}
	v := validation.NewValidatorWithKeywords(strict, keywords)
	var results []validationResult

	if len(args) > 0 {
		// Single item validation
		name := args[0]
		result, err := validateSingleItem(v, openspecDir, name, itemType)
		if err != nil {
			return err
		}
		results = append(results, *result)
	} else if all || changesOnly || specsOnly {
		// Batch validation
		var mu sync.Mutex
		var wg sync.WaitGroup

		if all || changesOnly {
			changes := utils.GetActiveChangeIDs(projectRoot)
			for _, id := range changes {
				wg.Add(1)
				go func(changeID string) {
					defer wg.Done()
					proposalPath := filepath.Join(openspecDir, "changes", changeID, "proposal.md")
					report := v.ValidateChange(proposalPath)

					// Also validate delta specs
					changeDir := filepath.Join(openspecDir, "changes", changeID)
					deltaReport := v.ValidateChangeDeltaSpecs(changeDir)

					// Merge reports
					mergedReport := mergeReports(report, deltaReport)

					mu.Lock()
					results = append(results, validationResult{
						Name:   changeID,
						Type:   "change",
						Report: mergedReport,
					})
					mu.Unlock()
				}(id)
			}
		}

		if all || specsOnly {
			specs := utils.GetSpecIDs(projectRoot)
			for _, id := range specs {
				wg.Add(1)
				go func(specID string) {
					defer wg.Done()
					specPath := filepath.Join(openspecDir, "specs", specID, "spec.md")
					report := v.ValidateSpec(specPath)
					mu.Lock()
					results = append(results, validationResult{
						Name:   specID,
						Type:   "spec",
						Report: report,
					})
					mu.Unlock()
				}(id)
			}
		}

		wg.Wait()
	} else {
		return fmt.Errorf("specify an item name, or use --all, --changes, or --specs")
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// Print results
	hasErrors := false
	for _, r := range results {
		printValidationResult(r)
		if !r.Report.Valid {
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}

func validateSingleItem(v *validation.Validator, openspecDir, name, itemType string) (*validationResult, error) {
	if itemType == "" {
		changePath := filepath.Join(openspecDir, "changes", name, "proposal.md")
		specPath := filepath.Join(openspecDir, "specs", name, "spec.md")

		if utils.FileExists(changePath) {
			itemType = "change"
		} else if utils.FileExists(specPath) {
			itemType = "spec"
		} else {
			return nil, fmt.Errorf("item %q not found as change or spec", name)
		}
	}

	if itemType == "change" {
		proposalPath := filepath.Join(openspecDir, "changes", name, "proposal.md")
		report := v.ValidateChange(proposalPath)
		changeDir := filepath.Join(openspecDir, "changes", name)
		deltaReport := v.ValidateChangeDeltaSpecs(changeDir)
		merged := mergeReports(report, deltaReport)
		return &validationResult{Name: name, Type: "change", Report: merged}, nil
	}

	specPath := filepath.Join(openspecDir, "specs", name, "spec.md")
	report := v.ValidateSpec(specPath)
	return &validationResult{Name: name, Type: "spec", Report: report}, nil
}

func mergeReports(a, b validation.Report) validation.Report {
	issues := append(a.Issues, b.Issues...)
	var errors, warnings, info int
	for _, issue := range issues {
		switch issue.Level {
		case validation.LevelError:
			errors++
		case validation.LevelWarning:
			warnings++
		case validation.LevelInfo:
			info++
		}
	}
	valid := errors == 0
	return validation.Report{
		Valid:  valid,
		Issues: issues,
		Summary: struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
			Info     int `json:"info"`
		}{
			Errors:   errors,
			Warnings: warnings,
			Info:     info,
		},
	}
}

func printValidationResult(r validationResult) {
	icon := color.GreenString("OK")
	if !r.Report.Valid {
		icon = color.RedString("FAIL")
	}

	fmt.Printf("\n%s %s (%s)\n", icon, r.Name, r.Type)

	for _, issue := range r.Report.Issues {
		var prefix string
		switch issue.Level {
		case validation.LevelError:
			prefix = color.RedString("  ERROR")
		case validation.LevelWarning:
			prefix = color.YellowString("  WARN ")
		case validation.LevelInfo:
			prefix = color.CyanString("  INFO ")
		}
		fmt.Printf("%s [%s] %s\n", prefix, issue.Path, issue.Message)
	}

	fmt.Printf("  Summary: %d errors, %d warnings, %d info\n",
		r.Report.Summary.Errors, r.Report.Summary.Warnings, r.Report.Summary.Info)
}
