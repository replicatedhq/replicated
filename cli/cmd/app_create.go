package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/integration"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

type createAppOpts struct {
	name string
}

func (r *runners) InitAppCreate(parent *cobra.Command) *cobra.Command {
	opts := createAppOpts{}
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new application",
		Long: `Create a new application in your Replicated account.

This command allows you to initialize a new application that can be distributed
and managed using the KOTS platform. When you create a new app, it will be set up
with default configurations, which you can later customize.

The NAME argument is required and will be used as the application's name.`,
		Example: `# Create a new app named "My App"
replicated app create "My App"

# Create a new app and output the result in JSON format
replicated app create "Another App" --output json

# Create a new app with a specific name and view details in table format
replicated app create "Custom App" --output table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if integrationTest != "" {
				ctx = context.WithValue(ctx, integration.IntegrationTestContextKey, integrationTest)
			}
			if logAPICalls != "" {
				ctx = context.WithValue(ctx, integration.APICallLogContextKey, logAPICalls)
			}

			if len(args) != 1 {
				return errors.New("missing app name")
			}
			opts.name = args[0]
			return r.createApp(ctx, cmd, opts, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) createApp(ctx context.Context, cmd *cobra.Command, opts createAppOpts, outputFormat string) error {
	kotsRestClient := kotsclient.VendorV3Client{
		HTTPClient: *r.platformAPI,
	}

	app, err := kotsRestClient.CreateKOTSApp(ctx, opts.name)
	if err != nil {
		return errors.Wrap(err, "create app")
	}

	apps := []types.AppAndChannels{
		{
			App: &types.App{
				ID:        app.Id,
				Name:      app.Name,
				Slug:      app.Slug,
				Scheduler: "kots",
			},
		},
	}

	return print.Apps(outputFormat, r.w, apps)
}
