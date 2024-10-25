package cmd

import (
	"context"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/integration"
	"github.com/spf13/cobra"
)

func (r *runners) IniReleaseList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all of an app's releases",
		Long:  "List all of an app's releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if integrationTest != "" {
				ctx = context.WithValue(ctx, integration.IntegrationTestContextKey, integrationTest)
			}
			if logAPICalls != "" {
				ctx = context.WithValue(ctx, integration.APICallLogContextKey, logAPICalls)
			}

			return r.releaseList(ctx, cmd, args)
		},
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
}

func (r *runners) releaseList(ctx context.Context, cmd *cobra.Command, args []string) error {
	releases, err := r.api.ListReleases(ctx, r.appID, r.appType)
	if err != nil {
		return err
	}

	return print.Releases(r.outputFormat, r.w, releases)
}
