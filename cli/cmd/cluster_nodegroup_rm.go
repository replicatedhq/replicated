package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rm",
		Short:        "delete a node group",
		Long:         ``,
		RunE:         r.rmClusterNodeGroup,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) rmClusterNodeGroup(cmd *cobra.Command, args []string) error {
	nodeGroupID := args[0]

	// get the nodegroup (we need the cluster id from it)
	nodeGroup, err := r.kotsAPI.GetClusterNodeGroup(nodeGroupID)
	if err != nil {
		return errors.Wrap(err, "get cluster node group")
	}

	nodeGroups, err := r.kotsAPI.RemoveClusterNodeGroup(nodeGroup)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "remove cluster node group")
	}

	return print.ClusterNodeGroups(r.outputFormat, r.w, nodeGroups, true)
}
