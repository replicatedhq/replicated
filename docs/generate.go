package main

import (
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
	cmd.RootCmd.SetOutput(ioutil.Discard)
	cmd.Execute(ioutil.Discard)

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
	cmdgen.GenerateCmdDocs(cmd.RootCmd, path, opts)
}
