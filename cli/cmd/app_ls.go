package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitAppList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls [NAME]",
		Short:        "list kots apps",
		Long:         `list kots apps, or a single app by name`,
		RunE:         r.listApps,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) listApps(_ *cobra.Command, args []string) error {
	kotsApps, err := r.kotsAPI.ListApps()
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
	return print.Apps(r.outputFormat, r.w, resultApps)
}
