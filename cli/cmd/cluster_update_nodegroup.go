package cmd

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpdateNodegroup(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup [ID]",
		Short: "Update a nodegroup for a test cluster.",
		Long: `The 'nodegroup' command allows you to update the configuration of a nodegroup within a test cluster. You can update attributes like the number of nodes, minimum and maximum node counts for autoscaling, and more.

If you do not provide the nodegroup ID, the command will try to resolve it based on the nodegroup name provided.`,
		Example: `# Update the number of nodes in a nodegroup
replicated cluster update nodegroup CLUSTER_ID --nodegroup-id NODEGROUP_ID --nodes 3

# Update the autoscaling limits for a nodegroup
replicated cluster update nodegroup CLUSTER_ID --nodegroup-id NODEGROUP_ID --min-nodes 2 --max-nodes 5`,
		RunE:              r.updateClusterNodegroup,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateClusterNodeGroupID, "nodegroup-id", "", "The ID of the nodegroup to update")
	cmd.Flags().StringVar(&r.args.updateClusterNodeGroupName, "nodegroup-name", "", "The name of the nodegroup to update")
	cmd.Flags().IntVar(&r.args.updateClusterNodeGroupCount, "nodes", 0, "The number of nodes in the nodegroup")
	cmd.Flags().StringVar(&r.args.updateClusterNodeGroupMinCount, "min-nodes", "", "The minimum number of nodes in the nodegroup")
	cmd.Flags().StringVar(&r.args.updateClusterNodeGroupMaxCount, "max-nodes", "", "The maximum number of nodes in the nodegroup")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide (default: table)")

	return cmd
}

func (r *runners) updateClusterNodegroup(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateClusterIDArg(args); err != nil {
		return errors.Wrap(err, "ensure cluster id arg")
	}

	// if we don't have a node group id, we need to get it
	nodeGroupID := r.args.updateClusterNodeGroupID
	if nodeGroupID == "" {
		cl, err := r.kotsAPI.GetCluster(r.args.updateClusterID)
		if err != nil {
			return errors.Wrap(err, "get cluster")
		}

		for _, ng := range cl.NodeGroups {
			if ng.Name == r.args.updateClusterNodeGroupName {
				nodeGroupID = ng.ID
				break
			}
		}
	}
	if nodeGroupID == "" {
		return errors.New("nodegroup not found")
	}

	opts := kotsclient.UpdateClusterNodegroupOpts{
		Count: int64(r.args.updateClusterNodeGroupCount),
	}

	if r.args.updateClusterNodeGroupMinCount != "" {
		parsed, err := strconv.ParseInt(r.args.updateClusterNodeGroupMinCount, 10, 64)
		if err != nil {
			return errors.New("min-nodes must be an integer if provided")
		}

		opts.MinCount = &parsed
	}

	if r.args.updateClusterNodeGroupMaxCount != "" {
		parsed, err := strconv.ParseInt(r.args.updateClusterNodeGroupMaxCount, 10, 64)
		if err != nil {
			return errors.New("max-nodes must be an integer if provided")
		}

		opts.MaxCount = &parsed
	}

	cl, ve, err := r.kotsAPI.UpdateClusterNodegroup(r.args.updateClusterID, nodeGroupID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "update cluster nodegroup")
	}

	if ve != nil && ve.Message != "" {
		return errors.New(ve.Message)
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}
