package schemas

import "embed"

//go:embed spec-driven/schema.yaml spec-driven/templates/proposal.md spec-driven/templates/spec.md spec-driven/templates/design.md spec-driven/templates/tasks.md
var BuiltinSchemas embed.FS
