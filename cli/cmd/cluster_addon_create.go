package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddonCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create cluster add-ons.",
		Long:  `Create new add-ons for a cluster. This command allows you to add functionality or services to a cluster by provisioning the required add-ons.`,
		Example: `  # Create a Postgres database add-on for a cluster
  replicated cluster addon create postgres CLUSTER_ID --version 13 --disk 100 --instance-type db.t3.micro

  # Create an object store bucket add-on for a cluster
  replicated cluster addon create object-store CLUSTER_ID --bucket-prefix mybucket

  # Create a Postgres add-on and wait for it to be ready
  replicated cluster addon create postgres CLUSTER_ID --version 13 --wait 5m

  # Perform a dry run for creating an object store add-on
  replicated cluster addon create object-store CLUSTER_ID --bucket-prefix mybucket --dry-run`,
	}
	parent.AddCommand(cmd)

	return cmd
}
