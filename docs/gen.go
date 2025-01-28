package main

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/spf13/cobra/doc"
)

const docsDest = "./gen/docs"

func main() {
	var stdin bytes.Buffer
	err := os.MkdirAll(docsDest, 0755)
	if err != nil {
		log.Fatal(err)
	}

	rootCmd := cmd.GetRootCmd()
	rootCmd.DisableAutoGenTag = true

	err = cmd.Execute(rootCmd, &stdin, io.Discard, io.Discard)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTree(rootCmd, docsDest)
	if err != nil {
		log.Fatal(err)
	}
}
