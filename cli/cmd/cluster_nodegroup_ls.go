package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls [ID_OR_NAME]",
		Aliases: []string{"list"},
		Short:   "List node groups for a cluster.",
		Long: `The 'cluster nodegroup ls' command lists all the node groups associated with a given cluster. Each node group defines a specific set of nodes with particular configurations, such as instance types and scaling options.

You can view information about the node groups within the specified cluster, including their ID, name, node count, and other configuration details.

You must provide the cluster ID or name to list its node groups.`,
		Example: `# List all node groups in a cluster with default table output
replicated cluster nodegroup ls CLUSTER_ID_OR_NAME

# List node groups with JSON output
replicated cluster nodegroup ls CLUSTER_ID_OR_NAME --output json

# List node groups with wide table output
replicated cluster nodegroup ls CLUSTER_ID_OR_NAME --output wide`,
		Args:              cobra.ExactArgs(1),
		RunE:              r.listNodeGroups,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) listNodeGroups(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("cluster id is required")
	}
	clusterID := args[0]

	cluster, err := r.kotsAPI.GetCluster(clusterID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "get cluster")
	}

	return print.NodeGroups(r.outputFormat, r.w, cluster.NodeGroups)
}
