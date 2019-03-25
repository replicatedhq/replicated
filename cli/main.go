package main

import (
	"fmt"
	"os"

	"github.com/replicatedhq/replicated/cli/cmd"
)

func main() {
	if err := cmd.Execute(nil, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
