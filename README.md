# OpenSpec-Go

[![CI](https://github.com/santif/openspec-go/actions/workflows/ci.yml/badge.svg)](https://github.com/santif/openspec-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fgist.githubusercontent.com%2Fsantif%2F9fd59b6fc015a2c756284d76a66a2d54%2Fraw%2Fcoverage.json)](https://github.com/santif/openspec-go/actions/workflows/ci.yml)
[![Go 1.26+](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Go Reference](https://pkg.go.dev/badge/github.com/santif/openspec-go.svg)](https://pkg.go.dev/github.com/santif/openspec-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

AI-native system for spec-driven development. OpenSpec manages **specs** (requirements documents with scenarios) and **changes** (proposals that modify specs through deltas). The CLI validates markdown-based specs and changes, tracks artifact completion via a dependency graph, and generates workflow instructions for 24+ AI coding tools.

## Features

- **Spec-driven workflow** — structured proposal → specs → design → tasks pipeline
- **Delta-based changes** — add, modify, remove, or rename requirements without touching main specs until archive
- **Artifact dependency graph** — topological ordering, completion detection, blocked-artifact tracking
- **Validation engine** — schema validation + semantic rules (SHALL/MUST keywords, scenario counts, length limits)
- **24 AI tool integrations** — generates skill/command files for Claude Code, Cursor, Windsurf, GitHub Copilot, and more
- **Custom schemas** — fork the built-in schema or create your own artifact workflows
- **Cross-platform** — Linux, macOS, Windows (amd64/arm64) via GoReleaser

## Quick Start

```bash
# Initialize openspec-go in your project
openspec init --tools claude

# Create a new change
openspec new change add-user-auth

# Edit the generated proposal
$EDITOR openspec/changes/add-user-auth/proposal.md

# Validate the change
openspec validate add-user-auth

# Check artifact progress
openspec status --change add-user-auth

# Archive when complete (applies deltas to main specs)
openspec archive add-user-auth
```

## Installation

### Quick install (Linux / macOS)

```bash
curl -sSL https://raw.githubusercontent.com/santif/openspec-go/main/install.sh | sh
```

### From source

```bash
git clone https://github.com/santif/openspec-go.git
cd openspec-go
make build
# Binary at bin/openspec
```

### Go install

```bash
go install github.com/santif/openspec-go/cmd/openspec@latest
```

### Pre-built binaries

Download from [Releases](https://github.com/santif/openspec-go/releases) for Linux, macOS, or Windows.

## Commands

### Core workflow

| Command | Description |
|---------|-------------|
| `openspec init [path]` | Initialize OpenSpec-Go in a project |
| `openspec new change <name>` | Create a new change directory with template |
| `openspec validate [name]` | Validate a change or spec |
| `openspec status` | Show artifact completion for a change |
| `openspec archive [name]` | Archive a change and apply deltas to main specs |
| `openspec list` | List changes (default) or specs |
| `openspec show [name]` | Display a change or spec (markdown or JSON) |
| `openspec view` | Interactive dashboard of specs and changes |
| `openspec instructions [artifact]` | Output enriched instructions for an artifact |

### Configuration

| Command | Description |
|---------|-------------|
| `openspec config list` | Show all global config values |
| `openspec config get <key>` | Get a config value (profile, delivery, workflows) |
| `openspec config set <key> <value>` | Set a config value |
| `openspec config unset <key>` | Reset a config value to default |
| `openspec config reset` | Reset all config to defaults |
| `openspec config edit` | Open config in `$EDITOR` |
| `openspec config path` | Show config file path |

### Schema management

| Command | Description |
|---------|-------------|
| `openspec schemas` | List available workflow schemas |
| `openspec schema which` | Show the active schema |
| `openspec schema validate [name]` | Validate a schema file |
| `openspec schema fork [name]` | Copy built-in schema to project for customization |

### Other

| Command | Description |
|---------|-------------|
| `openspec update [path]` | Regenerate AI tool instruction files |
| `openspec completion <shell>` | Generate shell completions (bash, zsh, fish, powershell) |
| `openspec feedback <message>` | Submit feedback via GitHub Issues |

### Common flags

- `--json` — JSON output (available on list, show, validate, status, schemas, instructions)
- `--no-color` — Disable color output (global)
- `--type change|spec` — Disambiguate when a name exists as both (show, validate)

## Configuration

### Global config

Stored at `~/.config/openspec/config.json` (XDG-compliant):

```json
{
  "profile": "core",
  "delivery": "both",
  "featureFlags": {}
}
```

- **profile** — `core` (propose, explore, apply, archive) or `custom` (all workflows)
- **delivery** — `skills`, `commands`, or `both`

### Project config

Stored at `openspec/config.yaml` in your project root:

```yaml
schema: spec-driven
profile: core
context: |
  This is a Go project using Chi router and PostgreSQL.
rules:
  proposal:
    - Keep proposals under 2 pages
  specs:
    - Use RFC 2119 keywords (SHALL, MUST)
```

- **schema** — which artifact workflow to use
- **context** — project context injected into instructions (max 50KB)
- **rules** — per-artifact rules mapped to artifact IDs

### Keyword Localization

OpenSpec-Go validates normative keywords (`SHALL`, `MUST`) in requirements and conditional keywords (`WHEN`, `THEN`, `AND`) in scenarios. Both can be customized in `openspec/config.yaml` to match your team's language:

```yaml
keywords:
  normative: ["DEBE", "DEBERÁ", "DEBERA"]
  conditionals:
    when: "CUANDO"
    then: "ENTONCES"
    and: "Y"
```

- **keywords.normative** — keywords required in spec requirements (default: `SHALL`, `MUST`)
- **keywords.conditionals** — keywords used in scenario steps (default: `WHEN`, `THEN`, `AND`)

## Project Data Layout

```
your-project/
  openspec/
    config.yaml              # Project configuration
    specs/
      <spec-name>/
        spec.md              # Main spec (requirements + scenarios)
    changes/
      <change-name>/
        .openspec.yaml       # Change metadata (schema, created date)
        proposal.md          # Proposal document
        specs/
          <spec-name>/
            spec.md          # Delta spec (ADDED/MODIFIED/REMOVED/RENAMED)
        design.md            # Technical design (optional)
        tasks.md             # Implementation checklist
    archive/
      <date>-<change-name>/  # Archived changes
    schemas/
      <schema-name>/         # Project-local schema overrides
        schema.yaml
        templates/
```

## Schema System

Schemas define artifact workflows as dependency DAGs. The default `spec-driven` schema:

```
proposal → specs → design → tasks → [apply]
                ↘          ↗
                  design
```

**Artifacts:**

| Artifact | Generates | Description |
|----------|-----------|-------------|
| proposal | `proposal.md` | WHY — problem/opportunity, what changes, impact |
| specs | `specs/**/*.md` | WHAT — requirements with scenarios (delta format) |
| design | `design.md` | HOW — technical decisions, architecture, trade-offs |
| tasks | `tasks.md` | DO — implementation checklist with checkboxes |

Schema resolution order: project-local (`openspec/schemas/`) → user override (`~/.local/share/openspec/`) → built-in (embedded).

Fork a schema to customize: `openspec schema fork spec-driven`

## AI Tool Integrations

OpenSpec-Go generates workflow skill/command files for these AI coding tools:

Amazon Q Developer, Antigravity, Auggie (Augment CLI), Claude Code, Cline, CodeBuddy Code, Codex, Continue, CoStrict, Crush, Cursor, Factory Droid, Gemini CLI, GitHub Copilot, iFlow, Kilo Code, Kiro, OpenCode, Pi, Qoder, Qwen Code, RooCode, Trae, Windsurf

Configure during init: `openspec init --tools claude,cursor` or `openspec init --tools all`

## Development

```bash
make build          # Build binary to bin/openspec
make test           # Run all tests with race detector
make vet            # Run go vet
make lint           # Run golangci-lint
make all            # vet + test + build
make install        # Build and copy to $GOPATH/bin
make clean          # Remove bin/
```

Requires Go 1.26+. Module: `github.com/santif/openspec-go`.

## License

MIT — see [LICENSE](LICENSE).

Copyright (c) 2026 Santiago Fernandez. Portions derived from [OpenSpec](https://github.com/Fission-AI/OpenSpec) (c) 2024 OpenSpec Contributors.
