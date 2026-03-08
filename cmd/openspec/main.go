package main

import (
	"os"

	"github.com/santif/openspec-go/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
