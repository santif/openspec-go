package utils

import (
	"os"
	"testing"
)

func TestIsInteractive_CIEnv(t *testing.T) {
	t.Setenv("CI", "true")
	if IsInteractive() {
		t.Error("expected IsInteractive() to return false when CI=true")
	}
}

func TestIsInteractive_NoCIEnv(t *testing.T) {
	t.Setenv("CI", "")
	// Without CI set, IsInteractive checks if stdin is a terminal via os.Stdin.Stat().
	// The result depends on the actual environment: true if stdin is a TTY, false otherwise.
	// We just verify it does not panic and returns a boolean consistent with the stdin state.
	got := IsInteractive()

	fi, err := os.Stdin.Stat()
	if err != nil {
		// If Stat fails, IsInteractive returns false
		if got {
			t.Error("expected false when os.Stdin.Stat() fails")
		}
		return
	}
	isTTY := fi.Mode()&os.ModeCharDevice != 0
	if got != isTTY {
		t.Errorf("IsInteractive() = %v, but stdin TTY check = %v", got, isTTY)
	}
}
