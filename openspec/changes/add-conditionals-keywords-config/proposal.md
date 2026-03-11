## Why

OpenSpec supports configurable normative keywords (SHALL/MUST → DEBE/DEBERA) but scenario conditional keywords (WHEN/THEN/AND) are hardcoded throughout validation messages, spec templates, skill templates, and schema instructions. Users who customize normative keywords for their language also need matching conditional keywords for consistent spec authoring.

## What Changes

- Add `keywords.conditionals` config with named keys (`when`, `then`, `and`) to project config
- Dynamically generate validation guide messages using configured conditional keywords
- Replace conditional keywords in spec templates when generating new changes
- Append a compact "Project Keywords" instruction block when generating skills/commands
- Append the same instruction block when loading schema instructions for artifact generation
- Add validation for the new conditionals config (non-empty strings, no regex metacharacters)

## Capabilities

### New Capabilities

_None — this extends existing capabilities._

### Modified Capabilities

- `project-config`: Add `conditionals` field to `KeywordsConfig` with named keys (when/then/and), parsing, and validation
- `validation-engine`: Use configured conditional keywords in guide messages (`GuideScenarioFormat`, `GuideMissingSpecSections`)
- `command-generation`: Inject conditional keywords instruction when generating skills/commands
- `instruction-generation`: Inject conditional keywords instruction when loading schema instructions

## Impact

- **Config**: `KeywordsConfig` struct gains `Conditionals *ConditionalsConfig` field
- **Validation**: `Validator` gains `conditionalKeywords` field; guide messages become dynamic
- **CLI**: `validate.go` and `new_change.go` thread conditionals config through
- **Command generation**: `GenerateForTool` and `SkillTemplate`/`CommandTemplate` receive conditionals config
- **Instruction loading**: `LoadEnrichedInstruction` receives conditionals config
- **Embedded templates**: Not modified — keyword adaptation happens at runtime
