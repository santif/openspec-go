package validation

import "fmt"

const (
	MinWhySectionLength      = 50
	MinPurposeLength         = 50
	MaxWhySectionLength      = 1000
	MaxRequirementTextLength = 500
	MaxDeltasPerChange       = 10
)

var Messages = struct {
	ScenarioEmpty              string
	RequirementEmpty           string
	RequirementNoShall         string
	RequirementNoScenarios     string
	SpecNameEmpty              string
	SpecPurposeEmpty           string
	SpecNoRequirements         string
	ChangeNameEmpty            string
	ChangeWhyTooShort          string
	ChangeWhyTooLong           string
	ChangeWhatEmpty            string
	ChangeNoDeltas             string
	ChangeTooManyDeltas        string
	DeltaSpecEmpty             string
	DeltaDescriptionEmpty      string
	PurposeTooBrief            string
	RequirementTooLong         string
	DeltaDescriptionTooBrief   string
	DeltaMissingRequirements   string
	GuideNoDeltas              string
	GuideMissingSpecSections   string
	GuideMissingChangeSections string
	GuideScenarioFormat        string
}{
	ScenarioEmpty:              "Scenario text cannot be empty",
	RequirementEmpty:           "Requirement text cannot be empty",
	RequirementNoShall:         "Requirement must contain SHALL or MUST keyword",
	RequirementNoScenarios:     "Requirement must have at least one scenario",
	SpecNameEmpty:              "Spec name cannot be empty",
	SpecPurposeEmpty:           "Purpose section cannot be empty",
	SpecNoRequirements:         "Spec must have at least one requirement",
	ChangeNameEmpty:            "Change name cannot be empty",
	ChangeWhyTooShort:          fmt.Sprintf("Why section must be at least %d characters", MinWhySectionLength),
	ChangeWhyTooLong:           fmt.Sprintf("Why section should not exceed %d characters", MaxWhySectionLength),
	ChangeWhatEmpty:            "What Changes section cannot be empty",
	ChangeNoDeltas:             "Change must have at least one delta",
	ChangeTooManyDeltas:        fmt.Sprintf("Consider splitting changes with more than %d deltas", MaxDeltasPerChange),
	DeltaSpecEmpty:             "Spec name cannot be empty",
	DeltaDescriptionEmpty:      "Delta description cannot be empty",
	PurposeTooBrief:            fmt.Sprintf("Purpose section is too brief (less than %d characters)", MinPurposeLength),
	RequirementTooLong:         fmt.Sprintf("Requirement text is very long (>%d characters). Consider breaking it down.", MaxRequirementTextLength),
	DeltaDescriptionTooBrief:   "Delta description is too brief",
	DeltaMissingRequirements:   "Delta should include requirements",
	GuideNoDeltas:              `No deltas found. Ensure your change has a specs/ directory with capability folders (e.g. specs/http-server/spec.md) containing .md files that use delta headers (## ADDED/MODIFIED/REMOVED/RENAMED Requirements) and that each requirement includes at least one "#### Scenario:" block. Tip: run "openspec change show <change-id> --json --deltas-only" to inspect parsed deltas.`,
	GuideMissingSpecSections:   "Missing required sections. Expected headers: \"## Purpose\" and \"## Requirements\". Example:\n## Purpose\n[brief purpose]\n\n## Requirements\n### Requirement: Clear requirement statement\nUsers SHALL ...\n\n#### Scenario: Descriptive name\n- **WHEN** ...\n- **THEN** ...",
	GuideMissingChangeSections: "Missing required sections. Expected headers: \"## Why\" and \"## What Changes\". Ensure deltas are documented in specs/ using delta headers.",
	GuideScenarioFormat:        "Scenarios must use level-4 headers. Convert bullet lists into:\n#### Scenario: Short name\n- **WHEN** ...\n- **THEN** ...\n- **AND** ...",
}
