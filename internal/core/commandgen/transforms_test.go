package commandgen

import "testing"

func TestTransformToHyphenCommands(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"single command", "Run /opsx:propose", "Run /opsx-propose"},
		{"multiple commands", "Use /opsx:apply and /opsx:archive", "Use /opsx-apply and /opsx-archive"},
		{"no commands", "No commands here", "No commands here"},
		{"empty string", "", ""},
		{"only prefix", "/opsx:", "/opsx-"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformToHyphenCommands(tc.input)
			if got != tc.expected {
				t.Errorf("TransformToHyphenCommands(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
