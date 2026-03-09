# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go port of the OpenSpec CLI — an AI-native system for spec-driven development. OpenSpec manages specs (requirements documents with scenarios) and changes (proposals that modify specs through deltas). The CLI validates markdown-based specs and changes, tracks artifact completion via a dependency graph, and supports multiple AI tool integrations.

## Build & Development Commands

```bash
make build          # Build binary to bin/openspec
make test           # Run all tests with race detector
make vet            # Run go vet
make lint           # Run golangci-lint (must be installed separately)
make all            # vet + test + build
go test ./internal/core/validation/...   # Run tests for a single package
go test -run TestName ./internal/...     # Run a specific test
```

The binary version is injected via ldflags: `-X github.com/santif/openspec-go/internal/cli.version=VERSION`

## Architecture

### Entry point & CLI layer
- `cmd/openspec/main.go` — calls `cli.Execute()`
- `internal/cli/` — Cobra commands. Each file registers its command in `init()` and adds it to `rootCmd`. `deprecated.go` provides backward-compat aliases (`change show/list/validate`, `spec show/list/validate`). `show` supports `--json` and `--type` flags.
- Version is set via the `version` var in `root.go` (ldflags)

### Core domain (`internal/core/`)

**Parsing pipeline** (`parsers/`):
- `markdown.go` — Generic markdown section parser (`ParseSections`). Builds a tree of `Section{Level, Title, Content, Children}`. All spec/change parsing builds on this.
- `ParseSpec()` expects `## Purpose` and `## Requirements` sections. Requirements are `###` children with `####` scenario subsections.
- `ParseChange()` expects `## Why` and `## What Changes` sections. Deltas parsed from bullet list: `- **SpecName:** description`
- `change.go` — `ParseChangeWithDeltas()` extends ParseChange by also reading `specs/<name>/spec.md` delta files from the filesystem
- `requirementblocks.go` — Parses delta spec files (ADDED/MODIFIED/REMOVED/RENAMED sections)
- `requirementsection.go` — `ExtractRequirementsSection()` structurally separates the `## Requirements` section into ordered blocks for delta operations

**Schema & artifact graph** (`artifactgraph/`):
- Schemas are YAML files defining artifacts with dependency DAGs. Each artifact has `id`, `generates` (glob), `template`, `requires` (list of artifact IDs).
- `schema.go` — Loads/parses schema YAML with validation (no duplicates, valid refs, no cycles via DFS)
- `graph.go` — `ArtifactGraph` provides topological sort (Kahn's algorithm), next-ready detection, blocked artifact detection
- `state.go` — `DetectCompleted()` checks filesystem for generated files using doublestar glob matching
- `resolver.go` — Schema resolution chain: user override (`~/.local/share/openspec/`) → project-local (`openspec/schemas/`) → built-in (embedded via `schemas/embed.go`)
- `instructionloader.go` — Loads instruction files referenced by artifacts

**Validation** (`validation/`):
- `validator.go` — Validates specs and changes. Schema validation (required fields) + semantic rules (SHALL/MUST keywords, scenario counts, length limits)
- `constants.go` — All validation messages and thresholds
- `types.go` — `Report{Valid, Issues, Summary}`, `Issue{Level, Path, Message}`

**Command generation** (`commandgen/`):
- Adapter-based system that generates skill/command markdown files for multiple AI tools. `generator.go` dispatches to per-tool adapters via `ToolCommandAdapter` interface (defined in `types.go`). `registry.go` maps tool IDs to adapters (specialized ones for Claude, Cursor, Codex, Cline, OpenCode, Factory, Windsurf + generic adapter for the rest).
- `skilltemplates.go` / `transforms.go` — Template rendering and content transforms
- `embed.go` — `//go:embed` for skill/command markdown templates
- `yaml.go` — YAML frontmatter generation for tool-specific formats

**Spec application** (`specsapply/`):
- Applies delta operations (ADDED/MODIFIED/REMOVED/RENAMED) from change delta specs to main specs with atomic validation-then-write. `ApplySpecs()` orchestrates, `BuildUpdatedSpec()` applies per-spec.

**Legacy cleanup** (`legacycleanup/`):
- Detects and removes old slash-command directories and OpenSpec marker blocks from config files during upgrades.

**Migration** (`migration/`):
- One-time profile migration — scans installed workflows and sets up global config profile/delivery mode. `MigrateIfNeeded()` is the entry point.

**Profile drift** (`profiledrift/`):
- Detects when on-disk skill/command artifacts drift from desired workflow configuration, signaling need for `openspec update`.

**Converters** (`converters/`):
- `ConvertSpecToJSON()` / `ConvertChangeToJSON()` — markdown-to-JSON conversion for `show --json`.

**Other core packages**:
- `schemas/types.go` — Domain types: `Spec`, `Change`, `Delta`, `Requirement`, `Scenario`, `DeltaOperation` (ADDED/MODIFIED/REMOVED/RENAMED)
- `config/config.go` — Constants (`OpenSpecDirName = "openspec"`, `OpenSpecMarkers`), AI tool definitions (with skill directory mappings), and `WorkflowToSkillDir` map (workflow→skill-dir mappings)
- `globalconfig/` — XDG-compliant global config (`~/.config/openspec/config.json`): profiles (core/custom), delivery mode, feature flags
- `projectconfig/` — Per-project `openspec/config.yaml`: schema, profile, workflows, context, rules (mapped to artifact IDs)
- `profiles/` — Workflow profiles: core workflows (`propose`, `explore`, `apply`, `archive`) and all workflows

### Utilities (`internal/utils/`)
- `filesystem.go` — File/dir helpers (EnsureDir, FileExists, ReadFile, WriteFile)
- `itemdiscovery.go` — Discovers changes/specs by scanning `openspec/changes/` and `openspec/specs/`
- `changemetadata.go` — Reads/writes `.openspec.yaml` per change directory
- `changename.go` — Kebab-case validation for change names
- `interactive.go` — Terminal prompts
- `shelldetection.go` — Detect user's shell
- `taskprogress.go` — Progress display utilities

### Embedded schemas (`schemas/`)
- `embed.go` — `//go:embed` for built-in schema files
- `spec-driven/schema.yaml` — Default schema defining artifacts: proposal, spec, design, tasks
- `spec-driven/templates/` — Markdown templates for each artifact

### Embedded command templates (`commandgen/templates/`)
- Embedded markdown templates for skill and command files (one per workflow, in `skills/` and `commands/` subdirectories)

## Key Patterns

- Tests cover most core packages and utils. Run `make test` or target specific packages.
- All CLI commands use `init()` for registration with Cobra, not a central command tree
- Batch validation runs concurrently with goroutines + sync.Mutex
- Module path is `github.com/santif/openspec-go`
- Project data lives in `openspec/` directory within the user's project (changes in `openspec/changes/<name>/`, specs in `openspec/specs/<name>/`)
- GoReleaser handles cross-platform builds (linux/darwin/windows, amd64/arm64)
