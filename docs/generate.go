package main

import (
	"bytes"
	"flag"
	"io/ioutil"

	"github.com/replicatedhq/replicated/cli/cmd"
	"github.com/replicatedhq/replicated/docs/cmdgen"

	"github.com/spf13/hugo/hugolib"
)

func main() {
	var path string

	flag.StringVar(&path, "path", "/tmp", "Directory where the docs will be created")
	flag.Parse()

	/*
	 * Because subcommands are added to the root command inside Execute(), we
	 * first have to call Execute() to get docs for all subcommands.
	 */
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd := cmd.GetRootCmd()

	rootCmd.SetOutput(ioutil.Discard)
	cmd.Execute(rootCmd, nil, &stdout, &stderr)

	opts := cmdgen.CmdDocsGeneratorOptions{
		BasePath: "/reference/vendor-cli/",
		MetadataAdditional: map[string]interface{}{
			"categories": []string{"Reference"},
		},
		Menus: hugolib.PageMenus{
			"menu.main": &hugolib.MenuEntry{
				URL:        "/docs/reference/vendor-cli",
				Name:       "Replicated Vendor CLI",
				Menu:       "main",
				Identifier: "vendor-cli",
				Parent:     "/reference",
				Weight:     600,
			},
		},
	}
	cmdgen.GenerateCmdDocs(rootCmd, path, opts)
}
