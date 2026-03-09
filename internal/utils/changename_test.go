package utils

import (
	"testing"
)

func TestValidateChangeName_Valid(t *testing.T) {
	valid := []string{
		"add-auth",
		"a",
		"upgrade-to-v2",
		"fix-bug",
		"simple",
		"a1",
		"add-feature-123",
	}

	for _, name := range valid {
		if !ValidateChangeName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}
}

func TestValidateChangeName_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"Add-Auth",
		"has spaces",
		"has_underscores",
		"add--double-hyphens",
		"-starts-with-hyphen",
		"ends-with-hyphen-",
		"123-starts-with-number",
		"UPPERCASE",
		"camelCase",
	}

	for _, name := range invalid {
		if ValidateChangeName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}
