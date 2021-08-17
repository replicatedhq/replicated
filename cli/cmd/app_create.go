package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitAppCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create kots apps",
		Long:         `create kots apps`,
		RunE:         r.createApp,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) createApp(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing app name")
	}
	appName := args[0]

	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	app, err := kotsRestClient.CreateKOTSApp(appName)

	if err != nil {
		return errors.Wrap(err, "list apps")
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

	return print.Apps(r.w, apps)
}
