package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitAppList(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "ls [NAME]",
		Short: "List applications",
		Long: `List all applications in your Replicated account,
or search for a specific application by name or ID.

This command displays information about your applications, including their
names, IDs, and associated channels. If a NAME argument is provided, it will
filter the results to show only applications that match the given name or ID.

The output can be customized using the --output flag to display results in
either table or JSON format.`,
		Example: `  # List all applications
  replicated app ls

  # Search for a specific application by name
  replicated app ls "My App"

  # List applications and output in JSON format
  replicated app ls --output json

  # Search for an application and display results in table format
  replicated app ls "App Name" --output table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.listApps(cmd, args, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listApps(cmd *cobra.Command, args []string, outputFormat string) error {
	kotsApps, err := r.kotsAPI.ListApps(false)
	if err != nil {
		return errors.Wrap(err, "list apps")
	}

	if len(args) == 0 {
		return print.Apps(r.outputFormat, r.w, kotsApps)
	}

	appSearch := args[0]
	var resultApps []types.AppAndChannels
	for _, app := range kotsApps {
		if strings.Contains(app.App.ID, appSearch) || strings.Contains(app.App.Slug, appSearch) {
			resultApps = append(resultApps, app)
		}
	}
	return print.Apps(outputFormat, r.w, resultApps)
}
