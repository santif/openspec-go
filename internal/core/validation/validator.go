package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/santif/openspec-go/internal/core/parsers"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/core/schemas"
	"github.com/santif/openspec-go/internal/utils"
)

var defaultNormativeKeywords = []string{"SHALL", "MUST"}

type Validator struct {
	StrictMode          bool
	normativeKeywords   []string
	normativeRegex      *regexp.Regexp
	conditionalKeywords projectconfig.ConditionalsConfig
}

func NewValidator(strict bool) *Validator {
	return NewValidatorWithKeywords(strict, nil, nil)
}

func NewValidatorWithKeywords(strict bool, keywords []string, conditionals *projectconfig.ConditionalsConfig) *Validator {
	if len(keywords) == 0 {
		keywords = defaultNormativeKeywords
	}
	cond := projectconfig.DefaultConditionals()
	if conditionals != nil {
		if conditionals.When != "" {
			cond.When = conditionals.When
		}
		if conditionals.Then != "" {
			cond.Then = conditionals.Then
		}
		if conditionals.And != "" {
			cond.And = conditionals.And
		}
	}
	v := &Validator{
		StrictMode:          strict,
		normativeKeywords:   keywords,
		normativeRegex:      buildNormativeRegex(keywords),
		conditionalKeywords: cond,
	}
	return v
}

// buildNormativeRegex builds a regex that matches any of the given keywords
// using explicit context-based boundaries (not \b) for Unicode safety.
func buildNormativeRegex(keywords []string) *regexp.Regexp {
	// Sort longest-first so longer keywords match before shorter prefixes
	sorted := make([]string, len(keywords))
	copy(sorted, keywords)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})

	var escaped []string
	for _, kw := range sorted {
		escaped = append(escaped, regexp.QuoteMeta(kw))
	}
	pattern := `(?:^|[\s,;.!?()])(?:` + strings.Join(escaped, "|") + `)(?:$|[\s,;.!?()])`
	return regexp.MustCompile(pattern)
}

func (v *Validator) containsNormativeKeyword(text string) bool {
	return v.normativeRegex.MatchString(text)
}

func (v *Validator) requirementNoKeywordMessage() string {
	return "Requirement must contain " + strings.Join(v.normativeKeywords, " or ") + " keyword"
}

func (v *Validator) guideScenarioFormat() string {
	return fmt.Sprintf(GuideScenarioFormatTemplate,
		v.conditionalKeywords.When, v.conditionalKeywords.Then, v.conditionalKeywords.And)
}

func (v *Validator) guideMissingSpecSections() string {
	return fmt.Sprintf(GuideMissingSpecSectionsTemplate,
		v.conditionalKeywords.When, v.conditionalKeywords.Then)
}

func (v *Validator) ValidateSpec(filePath string) Report {
	var issues []Issue
	specName := extractNameFromPath(filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(err.Error()),
		})
		return v.createReport(issues)
	}

	spec, err := parsers.ParseSpec(specName, string(content))
	if err != nil {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(err.Error()),
		})
		return v.createReport(issues)
	}

	issues = append(issues, v.validateSpecSchema(spec)...)
	issues = append(issues, v.applySpecRulesWithConditionals(spec)...)

	return v.createReport(issues)
}

func (v *Validator) ValidateSpecContent(specName, content string) Report {
	var issues []Issue

	spec, err := parsers.ParseSpec(specName, content)
	if err != nil {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(err.Error()),
		})
		return v.createReport(issues)
	}

	issues = append(issues, v.validateSpecSchema(spec)...)
	issues = append(issues, v.applySpecRulesWithConditionals(spec)...)

	return v.createReport(issues)
}

func (v *Validator) ValidateChange(filePath string) Report {
	var issues []Issue
	changeName := extractNameFromPath(filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(err.Error()),
		})
		return v.createReport(issues)
	}

	changeDir := filepath.Dir(filePath)
	change, err := parsers.ParseChangeWithDeltas(changeName, string(content), changeDir)
	if err != nil {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(err.Error()),
		})
		return v.createReport(issues)
	}

	issues = append(issues, validateChangeSchema(change)...)
	issues = append(issues, applyChangeRules(change)...)

	return v.createReport(issues)
}

func (v *Validator) ValidateChangeDeltaSpecs(changeDir string) Report {
	var issues []Issue
	specsDir := filepath.Join(changeDir, "specs")
	totalDeltas := 0

	type emptySpec struct {
		path     string
		sections []string
	}
	var missingHeaderSpecs []string
	var emptySectionSpecs []emptySpec

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		// No specs dir is treated as no deltas
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(Messages.ChangeNoDeltas),
		})
		return v.createReport(issues)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		specName := entry.Name()
		specFile := filepath.Join(specsDir, specName, "spec.md")
		content, err := os.ReadFile(specFile)
		if err != nil {
			continue
		}

		plan := parsers.ParseDeltaSpec(string(content))
		entryPath := specName + "/spec.md"

		var sectionNames []string
		if plan.SectionPresence.Added {
			sectionNames = append(sectionNames, "## ADDED Requirements")
		}
		if plan.SectionPresence.Modified {
			sectionNames = append(sectionNames, "## MODIFIED Requirements")
		}
		if plan.SectionPresence.Removed {
			sectionNames = append(sectionNames, "## REMOVED Requirements")
		}
		if plan.SectionPresence.Renamed {
			sectionNames = append(sectionNames, "## RENAMED Requirements")
		}

		hasSections := len(sectionNames) > 0
		entryCount := len(plan.Added) + len(plan.Modified) + len(plan.Removed) + len(plan.Renamed)
		hasEntries := entryCount > 0

		if !hasEntries {
			if hasSections {
				emptySectionSpecs = append(emptySectionSpecs, emptySpec{path: entryPath, sections: sectionNames})
			} else {
				missingHeaderSpecs = append(missingHeaderSpecs, entryPath)
			}
		}

		addedNames := make(map[string]bool)
		modifiedNames := make(map[string]bool)
		removedNames := make(map[string]bool)
		renamedFrom := make(map[string]bool)
		renamedTo := make(map[string]bool)

		// Validate ADDED
		for _, block := range plan.Added {
			key := parsers.NormalizeRequirementName(block.Name)
			totalDeltas++
			if addedNames[key] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Duplicate requirement in ADDED: %q", block.Name)})
			} else {
				addedNames[key] = true
			}
			reqText := extractRequirementText(block.Raw)
			if reqText == "" {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("ADDED %q is missing requirement text", block.Name)})
			} else if !v.containsNormativeKeyword(reqText) {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("ADDED %q must contain %s", block.Name, strings.Join(v.normativeKeywords, " or "))})
			}
			if countScenarios(block.Raw) < 1 {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("ADDED %q must include at least one scenario", block.Name)})
			}
		}

		// Validate MODIFIED
		for _, block := range plan.Modified {
			key := parsers.NormalizeRequirementName(block.Name)
			totalDeltas++
			if modifiedNames[key] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Duplicate requirement in MODIFIED: %q", block.Name)})
			} else {
				modifiedNames[key] = true
			}
			reqText := extractRequirementText(block.Raw)
			if reqText == "" {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("MODIFIED %q is missing requirement text", block.Name)})
			} else if !v.containsNormativeKeyword(reqText) {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("MODIFIED %q must contain %s", block.Name, strings.Join(v.normativeKeywords, " or "))})
			}
			if countScenarios(block.Raw) < 1 {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("MODIFIED %q must include at least one scenario", block.Name)})
			}
		}

		// Validate REMOVED
		for _, name := range plan.Removed {
			key := parsers.NormalizeRequirementName(name)
			totalDeltas++
			if removedNames[key] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Duplicate requirement in REMOVED: %q", name)})
			} else {
				removedNames[key] = true
			}
		}

		// Validate RENAMED
		for _, pair := range plan.Renamed {
			fromKey := parsers.NormalizeRequirementName(pair.From)
			toKey := parsers.NormalizeRequirementName(pair.To)
			totalDeltas++
			if renamedFrom[fromKey] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Duplicate FROM in RENAMED: %q", pair.From)})
			} else {
				renamedFrom[fromKey] = true
			}
			if renamedTo[toKey] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Duplicate TO in RENAMED: %q", pair.To)})
			} else {
				renamedTo[toKey] = true
			}
		}

		// Cross-section conflicts
		for n := range modifiedNames {
			if removedNames[n] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Requirement present in both MODIFIED and REMOVED: %q", n)})
			}
			if addedNames[n] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Requirement present in both MODIFIED and ADDED: %q", n)})
			}
		}
		for n := range addedNames {
			if removedNames[n] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("Requirement present in both ADDED and REMOVED: %q", n)})
			}
		}
		for _, pair := range plan.Renamed {
			fromKey := parsers.NormalizeRequirementName(pair.From)
			toKey := parsers.NormalizeRequirementName(pair.To)
			if modifiedNames[fromKey] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("MODIFIED references old name from RENAMED. Use new header for %q", pair.To)})
			}
			if addedNames[toKey] {
				issues = append(issues, Issue{Level: LevelError, Path: entryPath, Message: fmt.Sprintf("RENAMED TO collides with ADDED for %q", pair.To)})
			}
		}
	}

	// Empty section specs
	for _, es := range emptySectionSpecs {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    es.path,
			Message: fmt.Sprintf("Delta sections %s were found, but no requirement entries parsed. Ensure each section includes at least one \"### Requirement:\" block (REMOVED may use bullet list syntax).", formatSectionList(es.sections)),
		})
	}
	for _, p := range missingHeaderSpecs {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    p,
			Message: "No delta sections found. Add headers such as \"## ADDED Requirements\" or move non-delta notes outside specs/.",
		})
	}

	if totalDeltas == 0 {
		issues = append(issues, Issue{
			Level:   LevelError,
			Path:    "file",
			Message: v.enrichTopLevelError(Messages.ChangeNoDeltas),
		})
	}

	return v.createReport(issues)
}

// Schema validation (replaces Zod)
func (v *Validator) validateSpecSchema(spec *schemas.Spec) []Issue {
	var issues []Issue
	if spec.Name == "" {
		issues = append(issues, Issue{Level: LevelError, Path: "name", Message: Messages.SpecNameEmpty})
	}
	if spec.Overview == "" {
		issues = append(issues, Issue{Level: LevelError, Path: "overview", Message: Messages.SpecPurposeEmpty})
	}
	if len(spec.Requirements) == 0 {
		issues = append(issues, Issue{Level: LevelError, Path: "requirements", Message: Messages.SpecNoRequirements})
	}
	for i, req := range spec.Requirements {
		if req.Text == "" {
			issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("requirements[%d].text", i), Message: Messages.RequirementEmpty})
		} else if !v.containsNormativeKeyword(req.Text) {
			issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("requirements[%d].text", i), Message: v.requirementNoKeywordMessage()})
		}
		if len(req.Scenarios) == 0 {
			issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("requirements[%d].scenarios", i), Message: Messages.RequirementNoScenarios})
		}
		for j, sc := range req.Scenarios {
			if sc.RawText == "" {
				issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("requirements[%d].scenarios[%d]", i, j), Message: Messages.ScenarioEmpty})
			}
		}
	}
	return issues
}

// applySpecRules is now a method on Validator to access conditional keywords.
func (v *Validator) applySpecRulesWithConditionals(spec *schemas.Spec) []Issue {
	var issues []Issue
	if len(spec.Overview) < MinPurposeLength {
		issues = append(issues, Issue{Level: LevelWarning, Path: "overview", Message: Messages.PurposeTooBrief})
	}
	for i, req := range spec.Requirements {
		if len(req.Text) > MaxRequirementTextLength {
			issues = append(issues, Issue{Level: LevelInfo, Path: fmt.Sprintf("requirements[%d]", i), Message: Messages.RequirementTooLong})
		}
		if len(req.Scenarios) == 0 {
			issues = append(issues, Issue{Level: LevelWarning, Path: fmt.Sprintf("requirements[%d].scenarios", i), Message: Messages.RequirementNoScenarios + ". " + v.guideScenarioFormat()})
		}
	}
	return issues
}

func validateChangeSchema(change *schemas.Change) []Issue {
	var issues []Issue
	if change.Name == "" {
		issues = append(issues, Issue{Level: LevelError, Path: "name", Message: Messages.ChangeNameEmpty})
	}
	if len(change.Why) < MinWhySectionLength {
		issues = append(issues, Issue{Level: LevelError, Path: "why", Message: Messages.ChangeWhyTooShort})
	}
	if len(change.Why) > MaxWhySectionLength {
		issues = append(issues, Issue{Level: LevelError, Path: "why", Message: Messages.ChangeWhyTooLong})
	}
	if change.WhatChanges == "" {
		issues = append(issues, Issue{Level: LevelError, Path: "whatChanges", Message: Messages.ChangeWhatEmpty})
	}
	if len(change.Deltas) == 0 {
		msg := Messages.ChangeNoDeltas + ". " + Messages.GuideNoDeltas
		issues = append(issues, Issue{Level: LevelError, Path: "deltas", Message: msg})
	}
	if len(change.Deltas) > MaxDeltasPerChange {
		issues = append(issues, Issue{Level: LevelError, Path: "deltas", Message: Messages.ChangeTooManyDeltas})
	}
	for i, d := range change.Deltas {
		if d.Spec == "" {
			issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("deltas[%d].spec", i), Message: Messages.DeltaSpecEmpty})
		}
		if d.Description == "" {
			issues = append(issues, Issue{Level: LevelError, Path: fmt.Sprintf("deltas[%d].description", i), Message: Messages.DeltaDescriptionEmpty})
		}
	}
	return issues
}

func applyChangeRules(change *schemas.Change) []Issue {
	var issues []Issue
	const minDeltaDescriptionLength = 10
	for i, delta := range change.Deltas {
		if len(delta.Description) < minDeltaDescriptionLength {
			issues = append(issues, Issue{Level: LevelWarning, Path: fmt.Sprintf("deltas[%d].description", i), Message: Messages.DeltaDescriptionTooBrief})
		}
		if (delta.Operation == schemas.DeltaAdded || delta.Operation == schemas.DeltaModified) && len(delta.Requirements) == 0 {
			issues = append(issues, Issue{Level: LevelWarning, Path: fmt.Sprintf("deltas[%d].requirements", i), Message: string(delta.Operation) + " " + Messages.DeltaMissingRequirements})
		}
	}
	return issues
}

// Helper functions
func extractNameFromPath(filePath string) string {
	posixPath := utils.ToPosixPath(filePath)
	parts := strings.Split(posixPath, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "specs" || parts[i] == "changes" {
			if i < len(parts)-1 {
				return parts[i+1]
			}
		}
	}
	fileName := parts[len(parts)-1]
	dotIndex := strings.LastIndex(fileName, ".")
	if dotIndex > 0 {
		return fileName[:dotIndex]
	}
	return fileName
}

func (v *Validator) enrichTopLevelError(baseMessage string) string {
	msg := strings.TrimSpace(baseMessage)
	if msg == Messages.ChangeNoDeltas {
		return msg + ". " + Messages.GuideNoDeltas
	}
	if strings.Contains(msg, "spec must have a Purpose section") || strings.Contains(msg, "spec must have a Requirements section") {
		return msg + ". " + v.guideMissingSpecSections()
	}
	if strings.Contains(msg, "change must have a Why section") || strings.Contains(msg, "change must have a What Changes section") {
		return msg + ". " + Messages.GuideMissingChangeSections
	}
	return msg
}

var scenarioHeaderRegex = regexp.MustCompile(`(?m)^####\s+`)

func countScenarios(blockRaw string) int {
	return len(scenarioHeaderRegex.FindAllString(blockRaw, -1))
}

func extractRequirementText(blockRaw string) string {
	lines := strings.Split(blockRaw, "\n")
	metadataRe := regexp.MustCompile(`^\*\*[^*]+\*\*:`)
	scenarioRe := regexp.MustCompile(`^####\s+`)

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if scenarioRe.MatchString(line) {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if metadataRe.MatchString(trimmed) {
			continue
		}
		return trimmed
	}
	return ""
}

func formatSectionList(sections []string) string {
	if len(sections) == 0 {
		return ""
	}
	if len(sections) == 1 {
		return sections[0]
	}
	head := sections[:len(sections)-1]
	last := sections[len(sections)-1]
	return strings.Join(head, ", ") + " and " + last
}

func (v *Validator) createReport(issues []Issue) Report {
	var errors, warnings, info int
	for _, issue := range issues {
		switch issue.Level {
		case LevelError:
			errors++
		case LevelWarning:
			warnings++
		case LevelInfo:
			info++
		}
	}

	valid := errors == 0
	if v.StrictMode {
		valid = errors == 0 && warnings == 0
	}

	return Report{
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
