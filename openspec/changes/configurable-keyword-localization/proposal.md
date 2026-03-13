# Proposal: Configurable Keyword Localization

## Intent

OpenSpec validates normative keywords (`SHALL`, `MUST`) in requirements and conditional keywords (`WHEN`, `THEN`, `AND`) in scenarios, but both sets are hardcoded in English. Teams writing specs in other languages cannot pass validation without mixing English keywords into their native-language requirements and scenarios. Adding a `keywords` section to `openspec/config.yaml` lets projects define their own normative and conditional keywords while keeping the current English defaults.

## Scope

In scope:
- `keywords.normative` list field in project config for custom normative keywords
- `keywords.conditionals` map field in project config with named keys (`when`, `then`, `and`)
- Dynamic regex construction for normative keyword validation using configured values
- Dynamic validation guide messages using configured conditional keywords
- Keyword instruction block injection into generated skills/commands and schema instructions
- Config validation: warn on empty keywords, regex metacharacters, and empty conditional fields
- Partial configuration support: unconfigured fields fall back to English defaults

Out of scope:
- Localization of OpenSpec CLI output messages or error text
- Localization of section headers (`## Purpose`, `## Requirements`, etc.)
- Per-spec or per-change keyword overrides (project-wide only)

## Approach

Add a `keywords` section to the existing `ProjectConfig` YAML structure with two sub-fields: `normative` (string list) and `conditionals` (named map with `when`/`then`/`and` keys). Thread the resolved keywords from project config through the CLI layer into the validator and instruction/command generators. The validator builds its normative regex dynamically from the configured list (sorted longest-first, with context-based Unicode-safe boundaries). Conditional keywords are injected into validation guide messages via `fmt.Sprintf` templates and appended as an instruction block to generated skills and schema instructions. All fields are optional — omitting the section entirely preserves the current English defaults.
