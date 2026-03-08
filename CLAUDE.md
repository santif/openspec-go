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
- `internal/cli/` — Cobra commands. Each file registers its command in `init()` and adds it to `rootCmd`. Commands: `init`, `new change`, `list`, `show`, `validate`, `status`, `view`, `archive`, `update`, `instructions`, `config`, `schema`, `feedback`
- Version is set via the `version` var in `root.go` (ldflags)

### Core domain (`internal/core/`)

**Parsing pipeline** (`parsers/`):
- `markdown.go` — Generic markdown section parser (`ParseSections`). Builds a tree of `Section{Level, Title, Content, Children}`. All spec/change parsing builds on this.
- `ParseSpec()` expects `## Purpose` and `## Requirements` sections. Requirements are `###` children with `####` scenario subsections.
- `ParseChange()` expects `## Why` and `## What Changes` sections. Deltas parsed from bullet list: `- **SpecName:** description`
- `change.go` — `ParseChangeWithDeltas()` extends ParseChange by also reading `specs/<name>/spec.md` delta files from the filesystem
- `requirementblocks.go` — Parses delta spec files (ADDED/MODIFIED/REMOVED/RENAMED sections)

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

**Other core packages**:
- `schemas/types.go` — Domain types: `Spec`, `Change`, `Delta`, `Requirement`, `Scenario`, `DeltaOperation` (ADDED/MODIFIED/REMOVED/RENAMED)
- `config/config.go` — Constants (`OpenSpecDirName = "openspec"`, `OpenSpecMarkers`) and AI tool definitions (24 tools with skill directory mappings)
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

## Key Patterns

- **No tests exist yet** — the project has no `_test.go` files
- All CLI commands use `init()` for registration with Cobra, not a central command tree
- Batch validation runs concurrently with goroutines + sync.Mutex
- Module path is `github.com/santif/openspec-go`
- Project data lives in `openspec/` directory within the user's project (changes in `openspec/changes/<name>/`, specs in `openspec/specs/<name>/`)
- GoReleaser handles cross-platform builds (linux/darwin/windows, amd64/arm64)
