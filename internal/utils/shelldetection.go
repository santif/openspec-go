package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DetectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}
	if runtime.GOOS == "windows" {
		if comspec := os.Getenv("COMSPEC"); comspec != "" {
			base := strings.ToLower(filepath.Base(comspec))
			if strings.Contains(base, "powershell") || strings.Contains(base, "pwsh") {
				return "powershell"
			}
			return "cmd"
		}
		return "powershell"
	}
	return "bash"
}
