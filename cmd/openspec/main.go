package main

import (
	"os"

	"github.com/Fission-AI/openspec-go/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
