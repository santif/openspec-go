package specsapply

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/santif/openspec-go/internal/core/parsers"
	"github.com/santif/openspec-go/internal/core/projectconfig"
	"github.com/santif/openspec-go/internal/core/validation"
	"github.com/santif/openspec-go/internal/utils"
)

var multiNewlineRe = regexp.MustCompile(`\n{3,}`)

// SpecUpdate represents a delta spec file that needs to be applied.
type SpecUpdate struct {
	Source string
	Target string
	Exists bool
}

// ApplyResult holds counts for a single capability.
type ApplyResult struct {
	Capability string
	Added      int
	Modified   int
	Removed    int
	Renamed    int
}

// ApplyOutput is the result of applying all delta specs.
type ApplyOutput struct {
	ChangeName   string
	Capabilities []ApplyResult
	Totals       struct {
		Added    int
		Modified int
		Removed  int
		Renamed  int
	}
	NoChanges bool
}

// ApplyOptions controls the behavior of ApplySpecs.
type ApplyOptions struct {
	DryRun         bool
	SkipValidation bool
	Silent         bool
}

// BuildSpecSkeleton returns template markdown for a new spec.
func BuildSpecSkeleton(specFolderName, changeName string) string {
	return fmt.Sprintf("# %s Specification\n\n## Purpose\nTBD - created by archiving change %s. Update Purpose after archive.\n\n## Requirements\n", specFolderName, changeName)
}

// FindSpecUpdates scans changeDir/specs/ for subdirs with spec.md.
func FindSpecUpdates(changeDir, mainSpecsDir string) ([]SpecUpdate, error) {
	var updates []SpecUpdate
	changeSpecsDir := filepath.Join(changeDir, "specs")

	entries, err := os.ReadDir(changeSpecsDir)
	if err != nil {
		return nil, nil //nolint:nilerr // No specs directory is not an error
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		specFile := filepath.Join(changeSpecsDir, entry.Name(), "spec.md")
		if !utils.FileExists(specFile) {
			continue
		}
		targetFile := filepath.Join(mainSpecsDir, entry.Name(), "spec.md")
		updates = append(updates, SpecUpdate{
			Source: specFile,
			Target: targetFile,
			Exists: utils.FileExists(targetFile),
		})
	}

	return updates, nil
}

// BuildUpdatedSpec applies delta operations to a spec and returns the rebuilt content.
func BuildUpdatedSpec(update SpecUpdate, changeName string) (string, ApplyResult, error) {
	// Read delta spec content
	changeContent, err := utils.ReadFile(update.Source)
	if err != nil {
		return "", ApplyResult{}, fmt.Errorf("failed to read delta spec: %w", err)
	}

	plan := parsers.ParseDeltaSpec(changeContent)
	specName := filepath.Base(filepath.Dir(update.Target))

	// Pre-validate duplicates within sections
	addedNames := make(map[string]bool)
	for _, add := range plan.Added {
		name := parsers.NormalizeRequirementName(add.Name)
		if addedNames[name] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - duplicate requirement in ADDED for header "### Requirement: %s"`, specName, add.Name)
		}
		addedNames[name] = true
	}

	modifiedNames := make(map[string]bool)
	for _, mod := range plan.Modified {
		name := parsers.NormalizeRequirementName(mod.Name)
		if modifiedNames[name] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - duplicate requirement in MODIFIED for header "### Requirement: %s"`, specName, mod.Name)
		}
		modifiedNames[name] = true
	}

	removedNamesSet := make(map[string]bool)
	for _, rem := range plan.Removed {
		name := parsers.NormalizeRequirementName(rem)
		if removedNamesSet[name] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - duplicate requirement in REMOVED for header "### Requirement: %s"`, specName, rem)
		}
		removedNamesSet[name] = true
	}

	renamedFromSet := make(map[string]bool)
	renamedToSet := make(map[string]bool)
	for _, r := range plan.Renamed {
		fromNorm := parsers.NormalizeRequirementName(r.From)
		toNorm := parsers.NormalizeRequirementName(r.To)
		if renamedFromSet[fromNorm] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - duplicate FROM in RENAMED for header "### Requirement: %s"`, specName, r.From)
		}
		if renamedToSet[toNorm] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - duplicate TO in RENAMED for header "### Requirement: %s"`, specName, r.To)
		}
		renamedFromSet[fromNorm] = true
		renamedToSet[toNorm] = true
	}

	// Pre-validate cross-section conflicts
	type conflict struct {
		name string
		a    string
		b    string
	}
	var conflicts []conflict

	for n := range modifiedNames {
		if removedNamesSet[n] {
			conflicts = append(conflicts, conflict{name: n, a: "MODIFIED", b: "REMOVED"})
		}
		if addedNames[n] {
			conflicts = append(conflicts, conflict{name: n, a: "MODIFIED", b: "ADDED"})
		}
	}
	for n := range addedNames {
		if removedNamesSet[n] {
			conflicts = append(conflicts, conflict{name: n, a: "ADDED", b: "REMOVED"})
		}
	}

	// Renamed interplay
	for _, r := range plan.Renamed {
		fromNorm := parsers.NormalizeRequirementName(r.From)
		toNorm := parsers.NormalizeRequirementName(r.To)
		if modifiedNames[fromNorm] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - when a rename exists, MODIFIED must reference the NEW header "### Requirement: %s"`, specName, r.To)
		}
		if addedNames[toNorm] {
			return "", ApplyResult{}, fmt.Errorf(
				`%s validation failed - RENAMED TO header collides with ADDED for "### Requirement: %s"`, specName, r.To)
		}
	}

	if len(conflicts) > 0 {
		c := conflicts[0]
		return "", ApplyResult{}, fmt.Errorf(
			`%s validation failed - requirement present in multiple sections (%s and %s) for header "### Requirement: %s"`, specName, c.a, c.b, c.name)
	}

	// Verify at least one operation exists
	hasAnyDelta := len(plan.Added) + len(plan.Modified) + len(plan.Removed) + len(plan.Renamed)
	if hasAnyDelta == 0 {
		sourceName := filepath.Base(filepath.Dir(update.Source))
		return "", ApplyResult{}, fmt.Errorf(
			"delta parsing found no operations for %s: provide ADDED/MODIFIED/REMOVED/RENAMED sections in change spec", sourceName)
	}

	// Load or create base target content
	var targetContent string
	isNewSpec := false

	if utils.FileExists(update.Target) {
		targetContent, err = utils.ReadFile(update.Target)
		if err != nil {
			return "", ApplyResult{}, fmt.Errorf("failed to read target spec: %w", err)
		}
	} else {
		// New spec: only ADDED allowed
		if len(plan.Modified) > 0 || len(plan.Renamed) > 0 {
			return "", ApplyResult{}, fmt.Errorf(
				"%s: target spec does not exist; only ADDED requirements are allowed for new specs, MODIFIED and RENAMED operations require an existing spec", specName)
		}
		if len(plan.Removed) > 0 {
			fmt.Printf("⚠️  Warning: %s - %d REMOVED requirement(s) ignored for new spec (nothing to remove).\n", specName, len(plan.Removed))
		}
		isNewSpec = true
		targetContent = BuildSpecSkeleton(specName, changeName)
	}

	// Extract requirements section and build name->block map
	parts := parsers.ExtractRequirementsSection(targetContent)

	// We need an ordered map. Use a map for lookups plus track insertion via original blocks.
	nameToBlock := make(map[string]parsers.RequirementBlock)
	for _, block := range parts.BodyBlocks {
		nameToBlock[parsers.NormalizeRequirementName(block.Name)] = block
	}

	// Apply operations in order: RENAMED → REMOVED → MODIFIED → ADDED

	// RENAMED
	for _, r := range plan.Renamed {
		from := parsers.NormalizeRequirementName(r.From)
		to := parsers.NormalizeRequirementName(r.To)
		if _, ok := nameToBlock[from]; !ok {
			return "", ApplyResult{}, fmt.Errorf(
				`%s RENAMED failed for header "### Requirement: %s" - source not found`, specName, r.From)
		}
		if _, ok := nameToBlock[to]; ok {
			return "", ApplyResult{}, fmt.Errorf(
				`%s RENAMED failed for header "### Requirement: %s" - target already exists`, specName, r.To)
		}
		block := nameToBlock[from]
		newHeader := "### Requirement: " + to
		rawLines := strings.Split(block.Raw, "\n")
		rawLines[0] = newHeader
		renamedBlock := parsers.RequirementBlock{
			HeaderLine: newHeader,
			Name:       to,
			Raw:        strings.Join(rawLines, "\n"),
		}
		delete(nameToBlock, from)
		nameToBlock[to] = renamedBlock
		// Update parts.BodyBlocks to reflect rename for order preservation
		for i, b := range parts.BodyBlocks {
			if parsers.NormalizeRequirementName(b.Name) == from {
				parts.BodyBlocks[i] = renamedBlock
				break
			}
		}
	}

	// REMOVED
	for _, name := range plan.Removed {
		key := parsers.NormalizeRequirementName(name)
		if _, ok := nameToBlock[key]; !ok {
			if !isNewSpec {
				return "", ApplyResult{}, fmt.Errorf(
					`%s REMOVED failed for header "### Requirement: %s" - not found`, specName, name)
			}
			continue
		}
		delete(nameToBlock, key)
	}

	// MODIFIED
	for _, mod := range plan.Modified {
		key := parsers.NormalizeRequirementName(mod.Name)
		if _, ok := nameToBlock[key]; !ok {
			return "", ApplyResult{}, fmt.Errorf(
				`%s MODIFIED failed for header "### Requirement: %s" - not found`, specName, mod.Name)
		}
		// Verify header line matches key
		rawFirstLine := strings.Split(mod.Raw, "\n")[0]
		m := parsers.RequirementHeaderRegex.FindStringSubmatch(rawFirstLine)
		if m == nil || parsers.NormalizeRequirementName(m[1]) != key {
			return "", ApplyResult{}, fmt.Errorf(
				`%s MODIFIED failed for header "### Requirement: %s" - header mismatch in content`, specName, mod.Name)
		}
		nameToBlock[key] = mod
	}

	// ADDED
	for _, add := range plan.Added {
		key := parsers.NormalizeRequirementName(add.Name)
		if _, ok := nameToBlock[key]; ok {
			return "", ApplyResult{}, fmt.Errorf(
				`%s ADDED failed for header "### Requirement: %s" - already exists`, specName, add.Name)
		}
		nameToBlock[key] = add
	}

	// Recompose: preserve original block order + append new blocks
	var keptOrder []parsers.RequirementBlock
	seen := make(map[string]bool)
	for _, block := range parts.BodyBlocks {
		key := parsers.NormalizeRequirementName(block.Name)
		if replacement, ok := nameToBlock[key]; ok {
			keptOrder = append(keptOrder, replacement)
			seen[key] = true
		}
	}
	// Append newly added blocks in their original order from plan.Added
	for _, add := range plan.Added {
		key := parsers.NormalizeRequirementName(add.Name)
		if !seen[key] {
			keptOrder = append(keptOrder, nameToBlock[key])
			seen[key] = true
		}
	}

	// Build requirements body
	var reqParts []string
	if preamble := strings.TrimSpace(parts.Preamble); preamble != "" {
		reqParts = append(reqParts, strings.TrimRight(parts.Preamble, " \t\n"))
	}
	for _, b := range keptOrder {
		reqParts = append(reqParts, b.Raw)
	}
	reqBody := strings.TrimRight(strings.Join(reqParts, "\n\n"), " \t\n")

	// Rebuild full content
	var resultParts []string
	if parts.Before != "" {
		resultParts = append(resultParts, strings.TrimRight(parts.Before, " \t\n"))
	}
	resultParts = append(resultParts, parts.HeaderLine)
	resultParts = append(resultParts, reqBody)
	resultParts = append(resultParts, parts.After)

	// Filter empty leading part
	if len(resultParts) > 0 && resultParts[0] == "" {
		resultParts = resultParts[1:]
	}

	rebuilt := strings.Join(resultParts, "\n")
	rebuilt = multiNewlineRe.ReplaceAllString(rebuilt, "\n\n")

	counts := ApplyResult{
		Capability: specName,
		Added:      len(plan.Added),
		Modified:   len(plan.Modified),
		Removed:    len(plan.Removed),
		Renamed:    len(plan.Renamed),
	}

	return rebuilt, counts, nil
}

// WriteUpdatedSpec writes the rebuilt spec to disk and prints a summary.
func WriteUpdatedSpec(update SpecUpdate, rebuilt string, counts ApplyResult, silent bool) error {
	targetDir := filepath.Dir(update.Target)
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	if err := utils.WriteFile(update.Target, rebuilt); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	if !silent {
		specName := filepath.Base(filepath.Dir(update.Target))
		fmt.Printf("Applying changes to openspec/specs/%s/spec.md:\n", specName)
		if counts.Added > 0 {
			fmt.Printf("  + %d added\n", counts.Added)
		}
		if counts.Modified > 0 {
			fmt.Printf("  ~ %d modified\n", counts.Modified)
		}
		if counts.Removed > 0 {
			fmt.Printf("  - %d removed\n", counts.Removed)
		}
		if counts.Renamed > 0 {
			fmt.Printf("  → %d renamed\n", counts.Renamed)
		}
	}

	return nil
}

// ApplySpecs applies all delta specs from a change to main specs.
// Atomic: prepares ALL specs before writing ANY.
func ApplySpecs(projectRoot, changeName string, opts ApplyOptions) (ApplyOutput, error) {
	changeDir := filepath.Join(projectRoot, "openspec", "changes", changeName)
	mainSpecsDir := filepath.Join(projectRoot, "openspec", "specs")

	// Verify change exists
	info, err := os.Stat(changeDir)
	if err != nil || !info.IsDir() {
		return ApplyOutput{}, fmt.Errorf("change %q not found", changeName)
	}

	// Find specs to update
	specUpdates, err := FindSpecUpdates(changeDir, mainSpecsDir)
	if err != nil {
		return ApplyOutput{}, err
	}

	if len(specUpdates) == 0 {
		return ApplyOutput{
			ChangeName: changeName,
			NoChanges:  true,
		}, nil
	}

	// Prepare all updates first (no writes)
	type prepared struct {
		update  SpecUpdate
		rebuilt string
		counts  ApplyResult
	}
	var preparedUpdates []prepared

	for _, update := range specUpdates {
		rebuilt, counts, err := BuildUpdatedSpec(update, changeName)
		if err != nil {
			return ApplyOutput{}, err
		}
		preparedUpdates = append(preparedUpdates, prepared{
			update:  update,
			rebuilt: rebuilt,
			counts:  counts,
		})
	}

	// Validate rebuilt specs unless skipped
	if !opts.SkipValidation {
		// Read project config for custom keywords
		var keywords []string
		if cfg := projectconfig.ReadProjectConfig(projectRoot); cfg != nil && cfg.Keywords != nil {
			keywords = cfg.Keywords.Normative
		}
		v := validation.NewValidatorWithKeywords(false, keywords)
		for _, p := range preparedUpdates {
			specName := filepath.Base(filepath.Dir(p.update.Target))
			report := v.ValidateSpecContent(specName, p.rebuilt)
			if !report.Valid {
				var errors []string
				for _, issue := range report.Issues {
					if issue.Level == validation.LevelError {
						errors = append(errors, fmt.Sprintf("  ✗ %s", issue.Message))
					}
				}
				return ApplyOutput{}, fmt.Errorf("validation errors in rebuilt spec for %s:\n%s", specName, strings.Join(errors, "\n"))
			}
		}
	}

	// Build results and write
	output := ApplyOutput{
		ChangeName: changeName,
	}

	for _, p := range preparedUpdates {
		if !opts.DryRun {
			if err := WriteUpdatedSpec(p.update, p.rebuilt, p.counts, opts.Silent); err != nil {
				return ApplyOutput{}, err
			}
		} else if !opts.Silent {
			specName := filepath.Base(filepath.Dir(p.update.Target))
			fmt.Printf("Would apply changes to openspec/specs/%s/spec.md:\n", specName)
			if p.counts.Added > 0 {
				fmt.Printf("  + %d added\n", p.counts.Added)
			}
			if p.counts.Modified > 0 {
				fmt.Printf("  ~ %d modified\n", p.counts.Modified)
			}
			if p.counts.Removed > 0 {
				fmt.Printf("  - %d removed\n", p.counts.Removed)
			}
			if p.counts.Renamed > 0 {
				fmt.Printf("  → %d renamed\n", p.counts.Renamed)
			}
		}

		output.Capabilities = append(output.Capabilities, p.counts)
		output.Totals.Added += p.counts.Added
		output.Totals.Modified += p.counts.Modified
		output.Totals.Removed += p.counts.Removed
		output.Totals.Renamed += p.counts.Renamed
	}

	return output, nil
}
