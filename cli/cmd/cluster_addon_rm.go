package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type clusterAddonRmArgs struct {
	id        string
	clusterID string
}

func (r *runners) InitClusterAddonRm(parent *cobra.Command) *cobra.Command {
	args := clusterAddonRmArgs{}

	cmd := &cobra.Command{
		Use:   "rm CLUSTER_ID --id ADDON_ID",
		Short: "Remove cluster add-on by ID.",
		Long: `The 'cluster addon rm' command allows you to remove a specific add-on from a cluster by specifying the cluster ID and the add-on ID.

This command is useful when you want to deprovision an add-on that is no longer needed or when troubleshooting issues related to specific add-ons. The add-on will be removed immediately, and you will receive confirmation upon successful removal.`,
		Example: `  # Remove an add-on with ID 'abc123' from cluster 'cluster456'
  replicated cluster addon rm cluster456 --id abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.clusterAddonRmRun(args)
		},
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	err := r.clusterAddonRmFlags(cmd, &args)
	if err != nil {
		panic(err)
	}

	return cmd
}

func (r *runners) clusterAddonRmFlags(cmd *cobra.Command, args *clusterAddonRmArgs) error {
	cmd.Flags().StringVar(&args.id, "id", "", "The ID of the cluster add-on to remove (required)")
	cmd.RegisterFlagCompletionFunc("id", r.completeClusterIDs)
	err := cmd.MarkFlagRequired("id")
	if err != nil {
		return err
	}
	return nil
}

func (r *runners) clusterAddonRmRun(args clusterAddonRmArgs) error {
	err := r.kotsAPI.DeleteClusterAddon(args.clusterID, args.id)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.w, "Add-on %s has been deleted\n", args.id)
	return err
}
