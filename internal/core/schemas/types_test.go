package schemas

import (
	"encoding/json"
	"testing"
)

// --- DeltaOperation values ---

func TestDeltaOperation_Values(t *testing.T) {
	tests := []struct {
		name     string
		op       DeltaOperation
		expected string
	}{
		{"ADDED", DeltaAdded, "ADDED"},
		{"MODIFIED", DeltaModified, "MODIFIED"},
		{"REMOVED", DeltaRemoved, "REMOVED"},
		{"RENAMED", DeltaRenamed, "RENAMED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.expected {
				t.Errorf("DeltaOperation %s = %q, want %q", tt.name, tt.op, tt.expected)
			}
		})
	}
}

func TestDeltaOperation_AllUnique(t *testing.T) {
	ops := []DeltaOperation{DeltaAdded, DeltaModified, DeltaRemoved, DeltaRenamed}
	seen := make(map[DeltaOperation]bool)
	for _, op := range ops {
		if seen[op] {
			t.Errorf("duplicate DeltaOperation: %q", op)
		}
		seen[op] = true
	}
}

// --- JSON roundtrip ---

func TestSpec_JSONRoundtrip(t *testing.T) {
	original := Spec{
		Name:     "test-spec",
		Overview: "A test specification",
		Requirements: []Requirement{
			{
				Text: "Must do something",
				Scenarios: []Scenario{
					{RawText: "When X then Y"},
				},
			},
		},
		Metadata: &Metadata{
			Version: "1.0",
			Format:  "markdown",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal Spec: %v", err)
	}

	var decoded Spec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Spec: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Overview != original.Overview {
		t.Errorf("Overview = %q, want %q", decoded.Overview, original.Overview)
	}
	if len(decoded.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(decoded.Requirements))
	}
	if decoded.Requirements[0].Text != original.Requirements[0].Text {
		t.Errorf("Requirement text = %q, want %q", decoded.Requirements[0].Text, original.Requirements[0].Text)
	}
	if len(decoded.Requirements[0].Scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(decoded.Requirements[0].Scenarios))
	}
	if decoded.Requirements[0].Scenarios[0].RawText != original.Requirements[0].Scenarios[0].RawText {
		t.Errorf("Scenario RawText = %q, want %q", decoded.Requirements[0].Scenarios[0].RawText, original.Requirements[0].Scenarios[0].RawText)
	}
	if decoded.Metadata == nil {
		t.Fatal("expected Metadata to be non-nil")
	}
	if decoded.Metadata.Version != "1.0" {
		t.Errorf("Metadata.Version = %q, want %q", decoded.Metadata.Version, "1.0")
	}
}

func TestChange_JSONRoundtrip(t *testing.T) {
	original := Change{
		Name:        "test-change",
		Why:         "Because reasons",
		WhatChanges: "Some things change",
		Deltas: []Delta{
			{
				Spec:        "my-spec",
				Operation:   DeltaAdded,
				Description: "New delta",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal Change: %v", err)
	}

	var decoded Change
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Change: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Why != original.Why {
		t.Errorf("Why = %q, want %q", decoded.Why, original.Why)
	}
	if decoded.WhatChanges != original.WhatChanges {
		t.Errorf("WhatChanges = %q, want %q", decoded.WhatChanges, original.WhatChanges)
	}
	if len(decoded.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(decoded.Deltas))
	}
	if decoded.Deltas[0].Spec != "my-spec" {
		t.Errorf("Delta.Spec = %q, want %q", decoded.Deltas[0].Spec, "my-spec")
	}
	if decoded.Deltas[0].Operation != DeltaAdded {
		t.Errorf("Delta.Operation = %q, want %q", decoded.Deltas[0].Operation, DeltaAdded)
	}
}
