# openspec-go

## 0.3.0

### Minor Changes

- [`8620e18`](https://github.com/santif/openspec-go/commit/8620e18bcd93349dac24cc4e45bd80779098e35b) feat: add configurable normative keywords validation

  Added support for configuring normative keywords (SHALL, MUST, etc.) in project config. Validators now respect project-level keyword overrides, allowing teams to customize which keywords are required in spec requirements and scenarios.

- [`bdfc459`](https://github.com/santif/openspec-go/commit/bdfc459493efd752f82cae7b4e6e67f707755e04) feat: add one-liner installer script for Linux and macOS

  New `install.sh` script supports both Linux and macOS with automatic architecture detection, GitHub release downloads, and POSIX-compatible installation to `/usr/local/bin`.

- [`de2bc91`](https://github.com/santif/openspec-go/commit/de2bc9108543e6809b8b602f13955186dbd4d0bc) ci: add GitHub Actions CI/CD and golangci-lint config

  Set up CI pipeline with Go test, vet, lint, and GoReleaser for automated releases on tags. Includes Codecov integration for tracking test coverage.

### Patch Changes

- [`99f4ab8`](https://github.com/santif/openspec-go/commit/99f4ab8027668050b58b873009176501d4c1823e) fix: convert 14 spec files from delta format to proper spec format

  Corrected embedded spec files that were incorrectly using delta format (ADDED/MODIFIED sections) instead of the standard spec format with Purpose and Requirements sections.

- [`f5c281c`](https://github.com/santif/openspec-go/commit/f5c281c2e85027e1b42e43452e2d19298682f41a) fix: show correct version when installed via `go install` ([#3](https://github.com/santif/openspec-go/pull/3))

  Fixed version display to show the Go module version when the binary is installed via `go install` instead of GoReleaser ldflags.

- [`ad41949`](https://github.com/santif/openspec-go/commit/ad41949c8a3a90242e2c6a7b039150fd2c876fde) fix: POSIX-compatible sed and graceful sudo fallback in installer ([#4](https://github.com/santif/openspec-go/pull/4))

  Fixed installer script to use POSIX-compatible `sed` syntax and gracefully handle systems where `sudo` is not available.

- [`8908bf3`](https://github.com/santif/openspec-go/commit/8908bf306a4ea4d9ecf061c83c67e09fb7cfa6c6) test: increase overall test coverage from 76% to 88.8% ([#1](https://github.com/santif/openspec-go/pull/1))

- [`63d6745`](https://github.com/santif/openspec-go/commit/63d6745fb5ec1c111578a29885c22aab1fd60467) test: improve test coverage across core packages ([#5](https://github.com/santif/openspec-go/pull/5))

- [`1b48cc1`](https://github.com/santif/openspec-go/commit/1b48cc1cbcdb93b789d380fb05388209c6541d59) test: improve coverage from 90.8% to 91.4% ([#6](https://github.com/santif/openspec-go/pull/6))

- [`daf012d`](https://github.com/santif/openspec-go/commit/daf012dad5512cb2ea8532153a6c077a9ebf17e9) ci: add Codecov integration and README badges ([#2](https://github.com/santif/openspec-go/pull/2))

- [`8ff0c79`](https://github.com/santif/openspec-go/commit/8ff0c793829fb2897fe4b4887dac57f57ed76a5c) docs: add normative keywords example to README project config ([#7](https://github.com/santif/openspec-go/pull/7))

## 0.2.0

### Minor Changes

- [`739ce9e`](https://github.com/santif/openspec-go/commit/739ce9e86981d0854f5dd56bd1b7875508d4dacb) feat: add test suite, new core packages, and commandgen overhaul

  Major expansion of core domain packages: added artifact graph with topological sort and dependency resolution, schema loading with cycle detection, spec application with atomic delta operations (ADDED/MODIFIED/REMOVED/RENAMED), legacy cleanup, migration, profile drift detection, and global config management. Overhauled command generation with per-tool adapters for 21+ AI tools.

- [`b3d8f97`](https://github.com/santif/openspec-go/commit/b3d8f976626e1dc9eca9b4531b0da18335fab9d9) feat: add OpenSpec project config, specs, and Claude skills/commands

  Added project configuration system (`openspec/config.yaml`) with schema, profile, workflow, and context settings. Included embedded spec definitions and Claude Code skill/command templates for the propose, explore, apply, and archive workflows.

### Patch Changes

- [`4b2095d`](https://github.com/santif/openspec-go/commit/4b2095d3eb6acdfdd2636eddbbc03b8ce2e7829b) docs: add project README

  Initial README with project overview, installation instructions, quick start guide, CLI reference, and architecture documentation.

- [`c8beac8`](https://github.com/santif/openspec-go/commit/c8beac8a20a293ef22260a384ca23efbd47411b7) docs: update CLAUDE.md to reflect current project state

## 0.1.0

### Minor Changes

- [`382673f`](https://github.com/santif/openspec-go/commit/382673f241810e08463119ec21626b614da719a3) feat: initial project scaffold

  Go port of the OpenSpec CLI. Set up project structure with Cobra CLI framework, core parsing pipeline (markdown, spec, change, delta parsers), validation engine with schema and semantic rules, command generation system with tool adapters, and markdown-to-JSON converters. Includes Makefile, GoReleaser config, and embedded schema/template assets.

- [`e7ed5c5`](https://github.com/santif/openspec-go/commit/e7ed5c5d2815ec0e39b7b7c97258b9af726fd5c2) feat: add LICENSE and migrate module to santif/openspec-go

  Added MIT license and migrated Go module path from `github.com/fission-ai/openspec-go` to `github.com/santif/openspec-go`.

- [`3c2cecd`](https://github.com/santif/openspec-go/commit/3c2cecd97fc7115c22ef1299cdde399071a3d2d2) feat: add skill/command generation for AI tools

  Added embedded markdown templates for skill and command files, supporting workflows (propose, explore, apply, archive) across multiple AI tool integrations.
