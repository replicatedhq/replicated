package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitDefaultShowCommand(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use: "show",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.showDefault(cmd, args[0], outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) showDefault(cmd *cobra.Command, defaultType string, outputFormat string) error {
	if cache.DefaultApp == "" {
		if outputFormat == "json" {
			fmt.Println("{}")
		} else {
			fmt.Println("No default app set")
		}
		return nil
	}

	app, err := getApp(cache.DefaultApp, r.api.KotsClient)
	if err != nil {
		return errors.Wrap(err, "get app")
	}

	if err := print.Apps(outputFormat, r.w, []types.AppAndChannels{{App: app}}); err != nil {
		return errors.Wrap(err, "print app")
	}

	return nil
}
