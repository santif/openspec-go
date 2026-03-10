## Why

The validation engine hardcodes `SHALL` and `MUST` as the only accepted normative keywords in requirement text. This blocks teams writing specs in languages other than English (e.g., Spanish: `DEBE`, `DEBERA`). Adding a `keywords.normative` field to `config.yaml` lets projects define their own normative keywords while keeping the current defaults for English projects.

## What Changes

- Add a `keywords` section to `openspec/config.yaml` with a `normative` list field
- Make the validator accept configurable normative keywords instead of the hardcoded `SHALL`/`MUST` regex
- Parametrize validation error messages to display the project's configured keywords
- Thread keywords from project config through the CLI to the validator

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `project-config`: Add `keywords.normative` field to the config structure and parsing
- `validation-engine`: Accept configurable normative keywords, build regex dynamically, parametrize error messages

## Impact

- `internal/core/projectconfig/projectconfig.go` — new `KeywordsConfig` struct and parsing
- `internal/core/validation/validator.go` — new constructor, dynamic regex, method-based keyword check
- `internal/core/validation/constants.go` — parametrized error message
- `internal/cli/validate.go` — read config and pass keywords to validator
- `internal/core/specsapply/specsapply.go` — pass keywords when creating validator for spec application
