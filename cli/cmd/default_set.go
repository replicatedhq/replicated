package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitDefaultSetCommand(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use: "set",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.setDefault(cmd, args[0], args[1], outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) setDefault(cmd *cobra.Command, defaultType string, defaultValue string, outputFormat string) error {
	app, err := getApp(defaultValue, r.api.KotsClient)
	if err != nil {
		return errors.Wrap(err, "get app")
	}

	if err := cache.SetDefault(defaultType, defaultValue); err != nil {
		return errors.Wrap(err, "set default in cache")
	}

	// print the default app
	if err := print.Apps(outputFormat, r.w, []types.AppAndChannels{{App: app}}); err != nil {
		return errors.Wrap(err, "print app")
	}

	return nil
}
