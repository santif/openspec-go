## Context

The validation engine hardcodes a `\b(SHALL|MUST)\b` regex to check normative keywords in spec requirements. This works for English but blocks projects writing specs in other languages. The project config (`openspec/config.yaml`) already supports extensible fields (schema, profile, workflows, context, rules) and the validator is instantiated in a small number of call sites.

## Goals / Non-Goals

**Goals:**
- Allow projects to configure custom normative keywords via `config.yaml`
- Maintain full backward compatibility — existing projects with no `keywords` config work identically
- Parametrize error messages to show the project's configured keywords

**Non-Goals:**
- Full i18n of section headers (## Purpose, ## Requirements, ## ADDED, etc.)
- Locale packs or language presets
- Internationalization of CLI error messages
- Configuring scenario labels (WHEN/THEN)

## Decisions

### 1. Config structure: nested `keywords.normative` vs flat `normativeKeywords`

**Chosen: nested `keywords.normative`**

```yaml
keywords:
  normative: ["DEBE", "DEBERA"]
```

Rationale: The `keywords` namespace leaves room for future keyword categories (e.g., `conditional: ["DEBERIA"]`, `scenario: { when: "CUANDO", then: "ENTONCES" }`) without polluting the top-level config. A flat `normativeKeywords` field would be simpler today but would require migration if we add more keyword types later.

### 2. Two constructors vs one constructor with optional keywords

**Chosen: two constructors**

- `NewValidator(strict bool)` — unchanged, uses defaults `["SHALL", "MUST"]`
- `NewValidatorWithKeywords(strict bool, keywords []string)` — accepts custom list

Rationale: Backward compatible. All existing call sites and tests continue to work without modification. The new constructor is only used where project config is available. `NewValidatorWithKeywords` with nil or empty keywords falls back to defaults.

### 3. Regex construction from keywords — Unicode-safe boundaries

Go's RE2 `\b` word boundary only recognizes ASCII word characters (`[0-9A-Za-z_]`). Accented characters like `Á` are not considered word characters, so `\b(DEBERÁ)\b` **fails to match**. This affects any language with diacritics (Spanish, French, Portuguese, etc.).

**Chosen: explicit context-based boundaries instead of `\b`**

```
keywords = ["DEBE", "DEBERÁ", "DEBERA"]

→ sorted longest-first: ["DEBERÁ", "DEBERA", "DEBE"]
→ regex: (?:^|[\s,;.!?()])(DEBERÁ|DEBERA|DEBE)(?:$|[\s,;.!?()])
```

Two key aspects:
- **Sort keywords longest-first** before building the alternation, so `DEBERÁ` matches before `DEBE` can match partially.
- **Use explicit boundaries** — start/end of string, whitespace, or common punctuation — instead of `\b`. This correctly handles Unicode characters while still preventing embedded matches (e.g., `NODEBE` won't match).

Keywords are escaped with `regexp.QuoteMeta` before joining. The regex is compiled once at validator construction time.

This approach also replaces the current `\b(SHALL|MUST)\b` for the defaults, making the boundary strategy consistent regardless of language.

### 4. `containsShallOrMust` becomes a method

The current package-level function `containsShallOrMust(text string) bool` becomes a method `(v *Validator) containsNormativeKeyword(text string) bool` that uses the validator's compiled regex. This keeps all keyword state on the validator struct.

### 5. Error message generation

`Messages.RequirementNoShall` (a static string) is replaced by a method `(v *Validator) requirementNoKeywordMessage() string` that formats the message with the configured keywords. The delta spec validation messages follow the same pattern.

## Risks / Trade-offs

- **[Risk] Keywords with accented characters** → Go RE2's `\b` does NOT handle Unicode. Mitigated by using explicit context-based boundaries (Decision 3).
- **[Risk] Empty keywords list configured** → Fallback to defaults with a warning. The validator always has at least one keyword.
- **[Risk] Multiple call sites need updating** → Only 2 call sites create validators: `cli/validate.go` and `specsapply/specsapply.go`. Both are straightforward.

## Open Questions

_(none — design decisions were explored in the preceding conversation)_
