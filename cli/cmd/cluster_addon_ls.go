package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

type clusterAddonLsArgs struct {
	clusterID    string
	outputFormat string
}

func (r *runners) InitClusterAddonLs(parent *cobra.Command) *cobra.Command {
	args := clusterAddonLsArgs{}

	cmd := &cobra.Command{
		Use:   "ls CLUSTER_ID",
		Short: "List cluster addons for a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.addonClusterLsRun(args)
		},
	}
	parent.AddCommand(cmd)

	_ = clusterAddonLsFlags(cmd, &args)

	return cmd
}

func clusterAddonLsFlags(cmd *cobra.Command, args *clusterAddonLsArgs) error {
	cmd.Flags().StringVar(&args.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	return nil
}

func (r *runners) addonClusterLsRun(args clusterAddonLsArgs) error {
	addons, err := r.kotsAPI.ListClusterAddons(args.clusterID)
	if err != nil {
		return err
	}

	return print.Addons(args.outputFormat, r.w, addons, true)
}
