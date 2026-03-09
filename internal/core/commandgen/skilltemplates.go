package commandgen

import "fmt"

// skillMetadata maps workflow IDs to their skill metadata.
var skillMetadata = map[string]struct {
	Name        string
	Description string
}{
	"propose":      {"openspec-propose", "Propose a new change with all artifacts generated in one step. Use when the user wants to quickly describe what they want to build and get a complete proposal with design, specs, and tasks ready for implementation."},
	"explore":      {"openspec-explore", "Enter explore mode - a thinking partner for exploring ideas, investigating problems, and clarifying requirements. Use when the user wants to think through something before or during a change."},
	"apply":        {"openspec-apply-change", "Implement tasks from an OpenSpec change. Use when the user wants to start implementing, continue implementation, or work through tasks."},
	"archive":      {"openspec-archive-change", "Archive a completed change in the experimental workflow. Use when the user wants to finalize and archive a change after implementation is complete."},
	"new":          {"openspec-new-change", "Start a new OpenSpec change using the experimental artifact workflow. Use when the user wants to create a new feature, fix, or modification with a structured step-by-step approach."},
	"continue":     {"openspec-continue-change", "Continue working on an OpenSpec change by creating the next artifact. Use when the user wants to progress their change, create the next artifact, or continue their workflow."},
	"ff":           {"openspec-ff-change", "Fast-forward through OpenSpec artifact creation. Use when the user wants to quickly create all artifacts needed for implementation without stepping through each one individually."},
	"sync":         {"openspec-sync-specs", "Sync delta specs from a change to main specs. Use when the user wants to update main specs with changes from a delta spec, without archiving the change."},
	"bulk-archive": {"openspec-bulk-archive-change", "Archive multiple completed changes at once. Use when archiving several parallel changes."},
	"verify":       {"openspec-verify-change", "Verify implementation matches change artifacts. Use when the user wants to validate that implementation is complete, correct, and coherent before archiving."},
	"onboard":      {"openspec-onboard", "Guided onboarding for OpenSpec - walk through a complete workflow cycle with narration and real codebase work."},
}

// commandMetadata maps workflow IDs to their command metadata.
var commandMetadata = map[string]struct {
	ID          string
	Name        string
	Description string
	Category    string
	Tags        []string
}{
	"propose":      {"propose", "OPSX: Propose", "Propose a new change - create it and generate all artifacts in one step", "Workflow", []string{"workflow", "artifacts", "experimental"}},
	"explore":      {"explore", "OPSX: Explore", "Enter explore mode - think through ideas, investigate problems, clarify requirements", "Workflow", []string{"workflow", "explore", "experimental", "thinking"}},
	"apply":        {"apply", "OPSX: Apply", "Implement tasks from an OpenSpec change (Experimental)", "Workflow", []string{"workflow", "artifacts", "experimental"}},
	"archive":      {"archive", "OPSX: Archive", "Archive a completed change in the experimental workflow", "Workflow", []string{"workflow", "archive", "experimental"}},
	"new":          {"new", "OPSX: New", "Start a new change using the experimental artifact workflow (OPSX)", "Workflow", []string{"workflow", "artifacts", "experimental"}},
	"continue":     {"continue", "OPSX: Continue", "Continue working on a change - create the next artifact (Experimental)", "Workflow", []string{"workflow", "artifacts", "experimental"}},
	"ff":           {"ff", "OPSX: Fast Forward", "Create a change and generate all artifacts needed for implementation in one go", "Workflow", []string{"workflow", "artifacts", "experimental"}},
	"sync":         {"sync", "OPSX: Sync", "Sync delta specs from a change to main specs", "Workflow", []string{"workflow", "specs", "experimental"}},
	"bulk-archive": {"bulk-archive", "OPSX: Bulk Archive", "Archive multiple completed changes at once", "Workflow", []string{"workflow", "archive", "experimental", "bulk"}},
	"verify":       {"verify", "OPSX: Verify", "Verify implementation matches change artifacts before archiving", "Workflow", []string{"workflow", "verify", "experimental"}},
	"onboard":      {"onboard", "OPSX: Onboard", "Guided onboarding - walk through a complete OpenSpec workflow cycle with narration", "Workflow", []string{"workflow", "onboarding", "tutorial", "learning"}},
}

// SkillTemplate returns the full skill template data for a given workflow.
func SkillTemplate(workflow string) SkillTemplateData {
	meta, ok := skillMetadata[workflow]
	if !ok {
		return SkillTemplateData{
			Name:         fmt.Sprintf("openspec-%s", workflow),
			Description:  fmt.Sprintf("OpenSpec %s workflow", workflow),
			Instructions: fmt.Sprintf("# %s\n\nRun `openspec instructions %s` for details.\n", workflow, workflow),
		}
	}

	instructions := loadSkillContent(workflow)
	if instructions == "" {
		instructions = fmt.Sprintf("# %s\n\nRun `openspec instructions %s` for details.\n", workflow, workflow)
	}

	return SkillTemplateData{
		Name:          meta.Name,
		Description:   meta.Description,
		Instructions:  instructions,
		License:       "MIT",
		Compatibility: "Requires openspec CLI.",
		Author:        "openspec",
		Version:       "1.0",
	}
}

// CommandTemplate returns the full command template data for a given workflow.
func CommandTemplate(workflow string) CommandTemplateData {
	meta, ok := commandMetadata[workflow]
	if !ok {
		return CommandTemplateData{
			ID:          workflow,
			Name:        fmt.Sprintf("OPSX: %s", workflow),
			Description: fmt.Sprintf("OpenSpec %s workflow", workflow),
			Body:        fmt.Sprintf("# %s\n\nRun `openspec instructions %s` for details.\n", workflow, workflow),
		}
	}

	body := loadCommandContent(workflow)
	if body == "" {
		body = fmt.Sprintf("# %s\n\nRun `openspec instructions %s` for details.\n", workflow, workflow)
	}

	return CommandTemplateData{
		ID:          meta.ID,
		Name:        meta.Name,
		Description: meta.Description,
		Category:    meta.Category,
		Tags:        meta.Tags,
		Body:        body,
	}
}

// GenerateSkillContent produces a complete skill file with YAML frontmatter.
func GenerateSkillContent(tmpl SkillTemplateData, cliVersion string) string {
	license := tmpl.License
	if license == "" {
		license = "MIT"
	}
	compatibility := tmpl.Compatibility
	if compatibility == "" {
		compatibility = "Requires openspec CLI."
	}
	author := tmpl.Author
	if author == "" {
		author = "openspec"
	}
	templateVersion := tmpl.Version
	if templateVersion == "" {
		templateVersion = "1.0"
	}

	return fmt.Sprintf(`---
name: %s
description: %s
license: %s
compatibility: %s
metadata:
  author: %s
  version: "%s"
  generatedBy: "%s"
---

%s`,
		EscapeYamlValue(tmpl.Name),
		EscapeYamlValue(tmpl.Description),
		license,
		EscapeYamlValue(compatibility),
		author,
		templateVersion,
		cliVersion,
		tmpl.Instructions,
	)
}
