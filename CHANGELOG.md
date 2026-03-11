# openspec-go

## 0.2.0

### Minor Changes

- [`3970612`](https://github.com/santif/openspec-go/commit/3970612) feat: add configurable conditional keywords (WHEN/THEN/AND) to project config ([#10](https://github.com/santif/openspec-go/pull/10))

  Added support for configuring conditional keywords (WHEN, THEN, AND) in project config alongside the existing normative keywords. Validators now respect project-level conditional keyword overrides, allowing teams to customize which conditional keywords are required in spec scenarios.

## 0.1.2

### Minor Changes

- [`bdfc459`](https://github.com/santif/openspec-go/commit/bdfc459493efd752f82cae7b4e6e67f707755e04) feat: add one-liner installer script for Linux and macOS

  New `install.sh` script supports both Linux and macOS with automatic architecture detection, GitHub release downloads, and POSIX-compatible installation to `/usr/local/bin`.

### Patch Changes

- [`ad41949`](https://github.com/santif/openspec-go/commit/ad41949c8a3a90242e2c6a7b039150fd2c876fde) fix: POSIX-compatible sed and graceful sudo fallback in installer ([#4](https://github.com/santif/openspec-go/pull/4))

  Fixed installer script to use POSIX-compatible `sed` syntax and gracefully handle systems where `sudo` is not available.

## 0.1.1

### Patch Changes

- [`f5c281c`](https://github.com/santif/openspec-go/commit/f5c281c2e85027e1b42e43452e2d19298682f41a) fix: show correct version when installed via `go install` ([#3](https://github.com/santif/openspec-go/pull/3))

  Fixed version display to show the Go module version when the binary is installed via `go install` instead of GoReleaser ldflags.

## 0.1.0

### Minor Changes

- [`382673f`](https://github.com/santif/openspec-go/commit/382673f241810e08463119ec21626b614da719a3) feat: initial project scaffold

  Go port of the OpenSpec CLI. Set up project structure with Cobra CLI framework, core parsing pipeline (markdown, spec, change, delta parsers), validation engine with schema and semantic rules, command generation system with tool adapters, and markdown-to-JSON converters. Includes Makefile, GoReleaser config, and embedded schema/template assets.

- [`e7ed5c5`](https://github.com/santif/openspec-go/commit/e7ed5c5d2815ec0e39b7b7c97258b9af726fd5c2) feat: add LICENSE and migrate module to santif/openspec-go

  Added MIT license and migrated Go module path from `github.com/fission-ai/openspec-go` to `github.com/santif/openspec-go`.

- [`3c2cecd`](https://github.com/santif/openspec-go/commit/3c2cecd97fc7115c22ef1299cdde399071a3d2d2) feat: add skill/command generation for AI tools

  Added embedded markdown templates for skill and command files, supporting workflows (propose, explore, apply, archive) across multiple AI tool integrations.

- [`739ce9e`](https://github.com/santif/openspec-go/commit/739ce9e86981d0854f5dd56bd1b7875508d4dacb) feat: add test suite, new core packages, and commandgen overhaul

  Major expansion of core domain packages: added artifact graph with topological sort and dependency resolution, schema loading with cycle detection, spec application with atomic delta operations (ADDED/MODIFIED/REMOVED/RENAMED), legacy cleanup, migration, profile drift detection, and global config management. Overhauled command generation with per-tool adapters for 21+ AI tools.

- [`b3d8f97`](https://github.com/santif/openspec-go/commit/b3d8f976626e1dc9eca9b4531b0da18335fab9d9) feat: add OpenSpec project config, specs, and Claude skills/commands

  Added project configuration system (`openspec/config.yaml`) with schema, profile, workflow, and context settings. Included embedded spec definitions and Claude Code skill/command templates for the propose, explore, apply, and archive workflows.

- [`8620e18`](https://github.com/santif/openspec-go/commit/8620e18bcd93349dac24cc4e45bd80779098e35b) feat: add configurable normative keywords validation

  Added support for configuring normative keywords (SHALL, MUST, etc.) in project config. Validators now respect project-level keyword overrides, allowing teams to customize which keywords are required in spec requirements and scenarios.

