package profiles

import (
	"testing"

	"github.com/santif/openspec-go/internal/core/globalconfig"
)

func TestCoreWorkflows_ContainsExpected(t *testing.T) {
	expected := []string{"propose", "explore", "apply", "archive"}

	if len(CoreWorkflows) != len(expected) {
		t.Fatalf("expected CoreWorkflows to have %d items, got %d", len(expected), len(CoreWorkflows))
	}

	for i, name := range expected {
		if CoreWorkflows[i] != name {
			t.Errorf("expected CoreWorkflows[%d] to be %q, got %q", i, name, CoreWorkflows[i])
		}
	}
}

func TestAllWorkflows_ContainsAll(t *testing.T) {
	if len(AllWorkflows) != 11 {
		t.Fatalf("expected AllWorkflows to have 11 items, got %d", len(AllWorkflows))
	}

	allSet := make(map[string]bool)
	for _, w := range AllWorkflows {
		allSet[w] = true
	}

	for _, cw := range CoreWorkflows {
		if !allSet[cw] {
			t.Errorf("expected AllWorkflows to contain core workflow %q", cw)
		}
	}
}

func TestGetProfileWorkflows_CoreProfile(t *testing.T) {
	result := GetProfileWorkflows(globalconfig.ProfileCore, nil)

	if len(result) != len(CoreWorkflows) {
		t.Fatalf("expected %d workflows, got %d", len(CoreWorkflows), len(result))
	}

	for i, name := range CoreWorkflows {
		if result[i] != name {
			t.Errorf("expected result[%d] to be %q, got %q", i, name, result[i])
		}
	}

	// Verify it returns a copy: modifying result must not affect CoreWorkflows
	result[0] = "modified"
	if CoreWorkflows[0] == "modified" {
		t.Error("expected GetProfileWorkflows to return a copy, but modifying result changed CoreWorkflows")
	}
}

func TestGetProfileWorkflows_CustomProfile(t *testing.T) {
	custom := []string{"propose", "apply", "verify"}
	result := GetProfileWorkflows(globalconfig.ProfileCustom, custom)

	if len(result) != len(custom) {
		t.Fatalf("expected %d workflows, got %d", len(custom), len(result))
	}

	for i, name := range custom {
		if result[i] != name {
			t.Errorf("expected result[%d] to be %q, got %q", i, name, result[i])
		}
	}
}

func TestGetProfileWorkflows_CustomProfileNilWorkflows(t *testing.T) {
	result := GetProfileWorkflows(globalconfig.ProfileCustom, nil)

	if result == nil {
		t.Fatal("expected non-nil slice for custom profile with nil workflows")
	}

	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}
