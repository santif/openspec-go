## 1. Project Config

- [x] 1.1 Add `KeywordsConfig` struct and `Keywords *KeywordsConfig` field to `ProjectConfig` in `projectconfig.go`
- [x] 1.2 Add parsing logic for `keywords.normative` in `ReadProjectConfig()`
- [x] 1.3 Add config validation: warn on empty normative list, reject regex metacharacters in keywords
- [x] 1.4 Add tests for parsing keywords config (present, absent, empty, with accents)

## 2. Validator

- [x] 2.1 Add `normativeKeywords []string` and `normativeRegex *regexp.Regexp` fields to `Validator` struct
- [x] 2.2 Add `NewValidatorWithKeywords(strict bool, keywords []string)` constructor that builds dynamic regex with `regexp.QuoteMeta`
- [x] 2.3 Update `NewValidator(strict bool)` to delegate to `NewValidatorWithKeywords` with nil keywords (defaults)
- [x] 2.4 Replace package-level `containsShallOrMust()` with method `(v *Validator) containsNormativeKeyword(text string) bool`
- [x] 2.5 Replace static `Messages.RequirementNoShall` with method `(v *Validator) requirementNoKeywordMessage() string`
- [x] 2.6 Update `validateSpecSchema()` and `ValidateChangeDeltaSpecs()` to use the new methods
- [x] 2.7 Add tests for custom keywords: valid match, no match, nil fallback, accented characters

## 3. CLI Integration

- [x] 3.1 Update `cli/validate.go` to read project config and pass keywords to `NewValidatorWithKeywords`
- [x] 3.2 Update `specsapply/specsapply.go` to pass keywords when creating the validator
- [x] 3.3 Add integration test: validate a spec with custom keywords configured in config.yaml
