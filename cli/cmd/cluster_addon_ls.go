package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

type clusterAddonLsArgs struct {
	clusterID string
}

func (r *runners) InitClusterAddonLs(parent *cobra.Command) *cobra.Command {
	args := clusterAddonLsArgs{}

	cmd := &cobra.Command{
		Use:     "ls CLUSTER_ID_OR_NAME",
		Aliases: []string{"list"},
		Short:   "List cluster add-ons for a cluster.",
		Long: `The 'cluster addon ls' command allows you to list all add-ons for a specific cluster. This command provides a detailed overview of the add-ons currently installed on the cluster, including their status and any relevant configuration details.

This can be useful for monitoring the health and configuration of add-ons or performing troubleshooting tasks.`,
		Example: `# List add-ons for a cluster with default table output
replicated cluster addon ls CLUSTER_ID_OR_NAME

# List add-ons for a cluster with JSON output
replicated cluster addon ls CLUSTER_ID_OR_NAME --output json

# List add-ons for a cluster with wide table output
replicated cluster addon ls CLUSTER_ID_OR_NAME --output wide`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.addonClusterLsRun(args.clusterID)
		},
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) addonClusterLsRun(clusterID string) error {
	addons, err := r.kotsAPI.ListClusterAddons(clusterID)
	if err != nil {
		return err
	}

	return print.Addons(r.outputFormat, r.w, addons, true)
}
