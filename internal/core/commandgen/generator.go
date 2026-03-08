package commandgen

import (
	"fmt"

	"github.com/santif/openspec-go/internal/core/globalconfig"
)

// GenerateForTool generates skill and/or command files for the given tool.
func GenerateForTool(toolID string, workflows []string, delivery globalconfig.Delivery) ([]CommandContent, error) {
	adapter := Get(toolID)
	if adapter == nil {
		return nil, fmt.Errorf("no adapter registered for tool %q", toolID)
	}

	var results []CommandContent

	switch delivery {
	case globalconfig.DeliverySkills:
		results = append(results, adapter.GenerateSkills(workflows)...)
	case globalconfig.DeliveryCommands:
		results = append(results, adapter.GenerateCommands(workflows)...)
	default: // DeliveryBoth
		results = append(results, adapter.GenerateSkills(workflows)...)
		results = append(results, adapter.GenerateCommands(workflows)...)
	}

	return results, nil
}
