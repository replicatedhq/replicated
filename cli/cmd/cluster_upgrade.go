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
		Use:   "upgrade [ID_OR_NAME]",
		Short: "Upgrade a test cluster.",
		Long:  `The 'upgrade' command upgrades a Kubernetes test cluster to a specified version. You must provide a cluster ID or name and the version to upgrade to. The upgrade can be simulated with a dry-run option, or you can choose to wait for the cluster to be fully upgraded.`,
		Example: `# Upgrade a cluster to a new Kubernetes version
replicated cluster upgrade CLUSTER_ID_OR_NAME --version 1.31

# Perform a dry run of a cluster upgrade without making any changes
replicated cluster upgrade CLUSTER_ID_OR_NAME --version 1.31 --dry-run

# Upgrade a cluster and wait for it to be ready
replicated cluster upgrade CLUSTER_ID_OR_NAME --version 1.31 --wait 30m`,
		Args:              cobra.ExactArgs(1),
		RunE:              r.upgradeCluster,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.upgradeClusterKubernetesVersion, "version", "", "Kubernetes version to upgrade to (format is distribution dependent)")
	cmd.Flags().BoolVar(&r.args.upgradeClusterDryRun, "dry-run", false, "Dry run")
	cmd.Flags().DurationVar(&r.args.upgradeClusterWaitDuration, "wait", 0, "Wait duration for cluster to be ready (leave empty to not wait)")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	_ = cmd.MarkFlagRequired("version")

	return cmd
}

func (r *runners) upgradeCluster(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("cluster id or name is required")
	}
	
	clusterID, err := r.getClusterIDFromArg(args[0])
	if err != nil {
		return errors.Wrap(err, "get cluster id from arg")
	}

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
