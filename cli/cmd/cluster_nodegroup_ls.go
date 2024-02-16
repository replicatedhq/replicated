package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls [ID]",
		Short: "List node groups for a cluster",
		Long:  `List node groups for a cluster`,
		Args:  cobra.ExactArgs(1),
		RunE:  r.listNodeGroups,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

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
