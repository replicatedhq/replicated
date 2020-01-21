package main

import (
	"os"

	"github.com/replicatedhq/replicated/cli/cmd"
)

func main() {
	if err := cmd.Execute(nil, os.Stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
