package utils

import (
	"testing"
)

func TestDetectShell_FromEnv(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := DetectShell()
	if got != "zsh" {
		t.Errorf("DetectShell() = %q, want %q", got, "zsh")
	}
}

func TestDetectShell_EmptyEnv(t *testing.T) {
	t.Setenv("SHELL", "")
	got := DetectShell()
	if got != "bash" {
		t.Errorf("DetectShell() = %q, want %q (default)", got, "bash")
	}
}

func TestDetectShell_PathVariants(t *testing.T) {
	t.Setenv("SHELL", "/usr/local/bin/fish")
	got := DetectShell()
	if got != "fish" {
		t.Errorf("DetectShell() = %q, want %q", got, "fish")
	}
}
