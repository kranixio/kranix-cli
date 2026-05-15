package main

import (
	"os"

	"github.com/kranix-io/kranix-cli/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
