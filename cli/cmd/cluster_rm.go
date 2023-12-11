package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ID [ID …]",
		Short: "Remove test clusters",
		Long: `Removes a cluster immediately.

You can specify the --all flag to terminate all clusters.`,
		RunE: r.removeCluster,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.removeClusterAll, "all", false, "remove all clusters")

	return cmd
}

func (r *runners) removeCluster(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.removeClusterAll {
		return errors.New("ID or --all flag required")
	} else if len(args) > 0 && r.args.removeClusterAll {
		return errors.New("cannot specify ID and --all flag")
	}

	if r.args.removeClusterAll {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}

		for _, cluster := range clusters {
			err := remove(r, cluster.ID)
			if err != nil {
				return errors.Wrap(err, "remove cluster")
			}
		}
	}

	for _, arg := range args {
		err := remove(r, arg)
		if err != nil {
			return errors.Wrap(err, "remove cluster")
		}
	}

	return nil
}

func remove(r *runners, clusterID string) error {
	err := r.kotsAPI.RemoveCluster(clusterID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "remove cluster")
	} else {
		fmt.Printf("removed cluster %s\n", clusterID)
	}
	return nil
}
