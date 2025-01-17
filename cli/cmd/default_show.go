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
		Use:   "show KEY",
		Short: "Show default value for a key",
		Long: `Shows defaul values for the specified key.

This command shows default values that will be used by other commands run by the current user.

Supported keys:
- app: the default application to use

The output can be customized using the --output flag to display results in
either table or JSON format.`,
		Example: `  # Show default application
  replicated default show app
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.showDefault(cmd, args[0], outputFormat)
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) showDefault(cmd *cobra.Command, defaultType string, outputFormat string) error {
	defaultValue, err := cache.GetDefault(defaultType)
	if err != nil {
		return errors.Wrap(err, "get default value")
	}

	if defaultValue == "" {
		if outputFormat == "json" {
			fmt.Println("{}")
		} else {
			fmt.Printf("No default set for %s\n", defaultType)
		}
		return nil
	}

	switch defaultType {
	case "app":
		app, err := getApp(defaultValue, r.api.KotsClient)
		if err != nil {
			return errors.Wrap(err, "get app")
		}

		if err := print.Apps(outputFormat, r.w, []types.AppAndChannels{{App: app}}); err != nil {
			return errors.Wrap(err, "print app")
		}
	}

	return nil
}
