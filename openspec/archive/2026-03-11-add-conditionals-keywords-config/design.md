## Context

OpenSpec already supports configurable normative keywords (`keywords.normative`) in project config. The pattern is established: `ProjectConfig` → `KeywordsConfig` → `Validator`. Conditional keywords (WHEN/THEN/AND) are currently hardcoded in validation messages, spec templates, skill templates, and schema instructions. This change extends the existing pattern to cover conditionals.

## Goals / Non-Goals

**Goals:**
- Allow projects to configure scenario conditional keywords (WHEN/THEN/AND) via `config.yaml`
- Adapted keywords flow into validation guide messages, spec templates, generated skills/commands, and schema instructions
- Follow the same patterns established by normative keywords config

**Non-Goals:**
- Validating that scenarios actually contain conditional keywords (no active enforcement)
- Full i18n of section headers (## Purpose, ## Requirements, ## ADDED, etc.)
- Modifying embedded template files — adaptation happens at runtime

## Decisions

### Decision 1: Named keys instead of a list

The `conditionals` field uses named keys (`when`, `then`, `and`) instead of a flat list like `normative`.

```yaml
keywords:
  conditionals:
    when: "CUANDO"
    then: "ENTONCES"
    and: "Y"
```

**Rationale:** Unlike normative keywords where any keyword in the list is equivalent (SHALL ≈ MUST), conditional keywords have distinct positional roles in scenarios. Named keys let the code know which keyword to place at each position in templates and examples.

**Alternative considered:** Flat list `["CUANDO", "ENTONCES", "Y"]` — rejected because the code cannot determine which keyword maps to which role without convention-based ordering, which is fragile.

### Decision 2: Runtime substitution, not template modification

Embedded templates (spec.md, sync.md, onboard.md, schema.yaml) remain unchanged with English WHEN/THEN/AND. Adaptation happens at runtime:
- **Spec template**: `strings.ReplaceAll` of `**WHEN**` → `**CUANDO**` etc. when writing to disk
- **Skill/command generation**: Append a compact instruction block
- **Schema instructions**: Append the same instruction block

**Rationale:** Embedded templates serve as readable documentation in the repository. Go's `//go:embed` doesn't support parameterization. Substitution is simple and predictable.

### Decision 3: Compact instruction injection for skills and schema

Instead of replacing WHEN/THEN in every example inside skill templates and schema instructions, a short block is appended:

```markdown
**Project Keywords**: This project uses custom scenario keywords. Use **CUANDO** instead of WHEN, **ENTONCES** instead of THEN, and **Y** instead of AND in all scenarios.
```

**Rationale:**
- Keeps embedded templates untouched and readable
- AI tools reliably follow explicit instructions
- One injection point rather than multiple find-and-replace operations across templates
- Works for both skill files (written to disk) and schema instructions (passed at runtime)

### Decision 4: Defaults and nil handling

When `conditionals` is nil (not configured), all behavior remains unchanged — hardcoded WHEN/THEN/AND in templates and messages. The `ConditionalsConfig` struct provides a `Resolve()` method returning effective keywords (configured or defaults).

**Rationale:** Zero-config projects see no change. Consistent with how `normative` handles nil (falls back to defaults).

### Decision 5: Validator receives conditionals alongside normative

Extend `NewValidatorWithKeywords` to accept a `ConditionalsConfig` parameter (or extend the Validator struct). The validator uses conditionals only for generating guide messages, not for active validation.

**Rationale:** Keeps the injection point minimal. The validator already generates dynamic messages for normative keywords — conditionals follow the same pattern.

## Risks / Trade-offs

- **Bold-pattern matching for spec template replacement** — Replacing `**WHEN**` relies on the template using bold markdown format. If template format changes, replacement breaks. → Mitigation: The replacement is a simple string match on a stable format; unit tests will catch regressions.
- **Instruction injection adds text to every generated skill** — Even a compact block adds bytes. → Mitigation: Only injected when `conditionals` is configured (non-nil); projects without it see no change.
- **Three conditional keywords may not cover all languages** — Some languages may want additional keywords (e.g., GIVEN). → Mitigation: Start with the three that map to WHEN/THEN/AND. The struct can be extended later with optional fields.

## Open Questions

_None — design is straightforward extension of existing patterns._
