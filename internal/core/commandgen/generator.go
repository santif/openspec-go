package commandgen

import (
	"fmt"
	"strings"

	"github.com/santif/openspec-go/internal/core/globalconfig"
	"github.com/santif/openspec-go/internal/core/projectconfig"
)

// GenerateForTool generates skill and/or command files for the given tool.
// If conditionals is non-nil, a "Project Keywords" instruction block is appended to each generated file.
func GenerateForTool(toolID string, workflows []string, delivery globalconfig.Delivery, version string, conditionals *projectconfig.ConditionalsConfig) ([]CommandContent, error) {
	adapter := Get(toolID)
	if adapter == nil {
		return nil, fmt.Errorf("no adapter registered for tool %q", toolID)
	}

	var results []CommandContent

	switch delivery {
	case globalconfig.DeliverySkills:
		results = append(results, adapter.GenerateSkills(workflows, version)...)
	case globalconfig.DeliveryCommands:
		results = append(results, adapter.GenerateCommands(workflows)...)
	default: // DeliveryBoth
		results = append(results, adapter.GenerateSkills(workflows, version)...)
		results = append(results, adapter.GenerateCommands(workflows)...)
	}

	if conditionals != nil {
		block := conditionalsInstructionBlock(conditionals)
		for i := range results {
			results[i].Content = strings.TrimRight(results[i].Content, "\n") + "\n\n" + block + "\n"
		}
	}

	return results, nil
}

// conditionalsInstructionBlock returns a compact instruction block for custom conditional keywords.
func conditionalsInstructionBlock(cond *projectconfig.ConditionalsConfig) string {
	return projectconfig.FormatConditionalsBlock(cond)
}
