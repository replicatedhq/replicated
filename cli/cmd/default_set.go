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
		Use:   "set KEY VALUE",
		Short: "Set default value for a key",
		Long: `Sets default value for the specified key.

This command sets default values that will be used by other commands run by the current user.

Supported keys:
- app: the default application to use

The output can be customized using the --output flag to display results in
either table or JSON format.`,
		Example: `  # Set default application
  replicated default set app my-app-slug`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.setDefault(cmd, args[0], args[1], outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) setDefault(cmd *cobra.Command, defaultType string, defaultValue string, outputFormat string) error {
	switch defaultType {
	case "app":
		app, err := getApp(defaultValue, r.api.KotsClient)
		if err != nil {
			return errors.Wrap(err, "get app")
		}

		if err := cache.SetDefault(defaultType, defaultValue); err != nil {
			return errors.Wrap(err, "set default in cache")
		}

		if err := print.Apps(outputFormat, r.w, []types.AppAndChannels{{App: app}}); err != nil {
			return errors.Wrap(err, "print app")
		}

		return nil
	default:
		return errors.Errorf("unknown default type: %s", defaultType)
	}
}
