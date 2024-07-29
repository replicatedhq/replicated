package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterUpgrade(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "upgrade [ID]",
		Short:             "Upgrade a test clusters",
		Long:              `Upgrade a test clusters`,
		Args:              cobra.ExactArgs(1),
		RunE:              r.upgradeCluster,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.upgradeClusterKubernetesVersion, "version", "", "Kubernetes version to upgrade to (format is distribution dependent)")
	cmd.Flags().BoolVar(&r.args.upgradeClusterDryRun, "dry-run", false, "Dry run")
	cmd.Flags().DurationVar(&r.args.upgradeClusterWaitDuration, "wait", 0, "Wait duration for cluster to be ready (leave empty to not wait)")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	_ = cmd.MarkFlagRequired("version")

	return cmd
}

func (r *runners) upgradeCluster(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("cluster id is required")
	}
	clusterID := args[0]

	opts := kotsclient.UpgradeClusterOpts{
		KubernetesVersion: r.args.upgradeClusterKubernetesVersion,
		DryRun:            r.args.upgradeClusterDryRun,
	}
	cl, err := r.upgradeAndWaitForCluster(clusterID, opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		_, err := fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

func (r *runners) upgradeAndWaitForCluster(clusterID string, opts kotsclient.UpgradeClusterOpts) (*types.Cluster, error) {
	cl, ve, err := r.kotsAPI.UpgradeCluster(clusterID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "upgrade cluster")
	}

	if ve != nil && len(ve.Errors) > 0 {
		if len(ve.SupportedDistributions) > 0 {
			_ = print.ClusterVersions("table", r.w, ve.SupportedDistributions)
		}
		return nil, fmt.Errorf("%s", errors.New(strings.Join(ve.Errors, ",")))
	}

	if opts.DryRun {
		return nil, nil
	}

	// if the wait flag was provided, we poll the api until the cluster is ready, or a timeout
	if r.args.upgradeClusterWaitDuration > 0 {
		return waitForCluster(r.kotsAPI, cl.ID, r.args.upgradeClusterWaitDuration)
	}

	return cl, nil
}
