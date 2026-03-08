package commandgen

import "fmt"

// SkillTemplate returns the markdown content for a given workflow skill.
func SkillTemplate(workflow string) string {
	switch workflow {
	case "propose":
		return proposeSkill
	case "explore":
		return exploreSkill
	case "apply":
		return applySkill
	case "archive":
		return archiveSkill
	case "new":
		return newSkill
	case "continue":
		return continueSkill
	case "ff":
		return ffSkill
	case "sync":
		return syncSkill
	case "bulk-archive":
		return bulkArchiveSkill
	case "verify":
		return verifySkill
	case "onboard":
		return onboardSkill
	default:
		return fmt.Sprintf("# %s\n\nRun `openspec instructions %s` for details.\n", workflow, workflow)
	}
}

const proposeSkill = `# Propose a Change

Create a new OpenSpec change proposal.

## Steps

1. Run ` + "`openspec new change <name>`" + ` to scaffold the change directory
2. Run ` + "`openspec instructions proposal --change <name>`" + ` to get enriched instructions
3. Follow the instructions to write the proposal document
4. Run ` + "`openspec validate --change <name>`" + ` to verify the proposal
`

const exploreSkill = `# Explore

Enter explore mode to think through ideas, investigate problems, and clarify requirements.

## Steps

1. Review the current specs with ` + "`openspec list specs`" + `
2. Use ` + "`openspec show <spec>`" + ` to examine existing specs
3. Use ` + "`openspec status`" + ` to understand the current state of changes
4. Discuss and refine requirements before creating a formal change
`

const applySkill = `# Apply a Change

Implement the tasks defined in a change.

## Steps

1. Run ` + "`openspec instructions apply --change <name>`" + ` to get implementation instructions
2. Run ` + "`openspec status --change <name>`" + ` to see which artifacts are pending
3. Implement the changes following the instructions
4. Run ` + "`openspec validate --change <name>`" + ` to verify the implementation
`

const archiveSkill = `# Archive a Change

Archive a completed change after implementation is verified.

## Steps

1. Run ` + "`openspec validate --change <name>`" + ` to ensure all artifacts pass
2. Run ` + "`openspec status --change <name>`" + ` to confirm all artifacts are complete
3. Run ` + "`openspec archive <name>`" + ` to archive the change
`

const newSkill = `# New Change

Create a new change from scratch with all artifacts generated in one step.

## Steps

1. Run ` + "`openspec new change <name>`" + ` to create the change
2. Fill in all required artifacts following the schema
3. Run ` + "`openspec validate --change <name>`" + ` to verify
`

const continueSkill = `# Continue a Change

Resume work on an existing change.

## Steps

1. Run ` + "`openspec status --change <name>`" + ` to see current progress
2. Run ` + "`openspec instructions <next-artifact> --change <name>`" + ` for the next pending artifact
3. Complete the artifact and validate
`

const ffSkill = `# Fast-Forward

Fast-forward a change by generating remaining artifacts.

## Steps

1. Run ` + "`openspec status --change <name>`" + ` to identify incomplete artifacts
2. Generate each remaining artifact using ` + "`openspec instructions <artifact> --change <name>`" + `
3. Validate the complete change
`

const syncSkill = `# Sync

Synchronize change artifacts with the latest spec state.

## Steps

1. Run ` + "`openspec list changes`" + ` to see active changes
2. Review each change's status with ` + "`openspec status --change <name>`" + `
3. Update artifacts as needed to match current specs
`

const bulkArchiveSkill = `# Bulk Archive

Archive multiple completed changes at once.

## Steps

1. Run ` + "`openspec list changes`" + ` to see all changes
2. Validate each change to confirm completeness
3. Archive completed changes with ` + "`openspec archive <name>`" + `
`

const verifySkill = `# Verify

Verify that a change's implementation matches its specification.

## Steps

1. Run ` + "`openspec status --change <name>`" + ` to check artifact completion
2. Run ` + "`openspec validate --change <name>`" + ` to validate all artifacts
3. Review any validation issues and fix them
`

const onboardSkill = `# Onboard

Get familiar with the project's OpenSpec setup.

## Steps

1. Run ` + "`openspec list specs`" + ` to see all specifications
2. Run ` + "`openspec list changes`" + ` to see active changes
3. Run ` + "`openspec show <spec>`" + ` to read individual specs
4. Run ` + "`openspec schema`" + ` to understand the artifact schema
`
