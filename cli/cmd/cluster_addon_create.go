package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddonCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create cluster add-ons.",
		Long:  `Create new add-ons for a cluster. This command allows you to add functionality or services to a cluster by provisioning the required add-ons.`,
		Example: `# Create an object store bucket add-on for a cluster
replicated cluster addon create object-store CLUSTER_ID --bucket-prefix mybucket

# Perform a dry run for creating an object store add-on
replicated cluster addon create object-store CLUSTER_ID --bucket-prefix mybucket --dry-run`,
	}
	parent.AddCommand(cmd)

	return cmd
}
