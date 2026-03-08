package profiles

import "github.com/santif/openspec-go/internal/core/globalconfig"

// CoreWorkflows are the minimal set of workflows available in the "core" profile.
var CoreWorkflows = []string{"propose", "explore", "apply", "archive"}

// AllWorkflows lists every known workflow.
var AllWorkflows = []string{
	"propose",
	"explore",
	"new",
	"continue",
	"apply",
	"ff",
	"sync",
	"archive",
	"bulk-archive",
	"verify",
	"onboard",
}

// GetProfileWorkflows returns the workflows enabled for the given profile.
func GetProfileWorkflows(profile globalconfig.Profile, customWorkflows []string) []string {
	if profile == globalconfig.ProfileCustom {
		if customWorkflows == nil {
			return []string{}
		}
		return customWorkflows
	}
	result := make([]string, len(CoreWorkflows))
	copy(result, CoreWorkflows)
	return result
}
