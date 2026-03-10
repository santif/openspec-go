## 1. Project Config

- [x] 1.1 Add `ConditionalsConfig` struct with `When`, `Then`, `And` fields to `internal/core/projectconfig/projectconfig.go`
- [x] 1.2 Add `Conditionals *ConditionalsConfig` field to `KeywordsConfig`
- [x] 1.3 Parse `keywords.conditionals` in `ReadProjectConfig` (extract when/then/and from YAML map)
- [x] 1.4 Add `ResolveConditionals()` method that returns effective keywords (configured or defaults)
- [x] 1.5 Extend `ValidateKeywords` to validate conditionals (non-empty strings, no regex metacharacters)
- [x] 1.6 Add tests for parsing, resolving, and validating conditionals config

## 2. Validation Engine

- [x] 2.1 Extend `Validator` struct to store conditional keywords (or a resolved `ConditionalsConfig`)
- [x] 2.2 Extend `NewValidatorWithKeywords` to accept conditionals parameter
- [x] 2.3 Make `GuideScenarioFormat` and `GuideMissingSpecSections` dynamic — generate from conditional keywords instead of using hardcoded constants
- [x] 2.4 Update `validate.go` CLI to read conditionals from project config and pass to validator
- [x] 2.5 Add tests for dynamic guide messages with custom and default conditionals

## 3. Spec Template Substitution

- [x] 3.1 In `new_change.go`, after reading the spec template, replace `**WHEN**`/`**THEN**`/`**AND**` with configured conditional keywords before writing to disk
- [x] 3.2 Add test for spec template keyword replacement

## 4. Command Generation

- [x] 4.1 Thread conditionals config into `GenerateForTool` and adapter interfaces
- [x] 4.2 In `SkillTemplate`/`CommandTemplate`, append "Project Keywords" instruction block when conditionals are configured
- [x] 4.3 Update `update` CLI command to read project config and pass conditionals to generation
- [x] 4.4 Add tests for skill/command generation with and without custom conditionals

## 5. Instruction Enrichment

- [x] 5.1 Thread conditionals config into `LoadEnrichedInstruction` and related functions
- [x] 5.2 Append "Project Keywords" instruction block to enriched instructions when conditionals are configured
- [x] 5.3 Update `instructions` CLI command to read project config and pass conditionals
- [x] 5.4 Add tests for instruction enrichment with and without custom conditionals
