package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type clusterAddonRmArgs struct {
	id           string
	clusterID    string
	outputFormat string
}

func (r *runners) InitClusterAddonRm(parent *cobra.Command) *cobra.Command {
	args := clusterAddonRmArgs{}

	cmd := &cobra.Command{
		Use:   "rm CLUSTER_ID --id ADDON_ID",
		Short: "Remove cluster addon by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.clusterAddonRmRun(args)
		},
	}
	parent.AddCommand(cmd)

	_ = clusterAddonRmFlags(cmd, &args)

	return cmd
}

func clusterAddonRmFlags(cmd *cobra.Command, args *clusterAddonRmArgs) error {
	cmd.Flags().StringVar(&args.id, "id", "", "The ID of the cluster addon to remove (required)")
	err := cmd.MarkFlagRequired("id")
	if err != nil {
		return err
	}
	cmd.Flags().StringVar(&args.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	return nil
}

func (r *runners) clusterAddonRmRun(args clusterAddonRmArgs) error {
	err := r.kotsAPI.DeleteClusterAddon(args.id)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(r.w, "Addon has been deleted")
	return err
}
