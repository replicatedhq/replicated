package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create test clusters",
		Long:         `Create test clusters`,
		RunE:         r.createCluster,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "Cluster name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "node-count", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "Cluster TTL (duration, max 48h)")
	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for cluster to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "The type of instance to use (e.g. x5.xlarge)")
	cmd.Flags()

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.MarkFlagRequired("distribution")
	cmd.MarkFlagRequired("version")

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	if r.args.createClusterName == "" {
		r.args.createClusterName = generateClusterName()
	}

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		NodeCount:              r.args.createClusterNodeCount,
		DiskGiB:                r.args.createClusterDiskGiB,
		TTL:                    r.args.createClusterTTL,
		DryRun:                 r.args.createClusterDryRun,
		InstanceType:           r.args.createClusterInstanceType,
	}
	cl, err := r.createAndWaitForCluster(opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		_, err := fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

func (r *runners) createAndWaitForCluster(opts kotsclient.CreateClusterOpts) (*types.Cluster, error) {
	cl, ve, err := r.kotsAPI.CreateCluster(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster")
	}

	if ve != nil && len(ve.Errors) > 0 {
		if len(ve.SupportedDistributions) > 0 {
			print.ClusterVersions("table", r.w, ve.SupportedDistributions)
		}
		return nil, fmt.Errorf("%s", errors.New(strings.Join(ve.Errors, ",")))
	}

	if opts.DryRun {
		return nil, nil
	}

	// if the wait flag was provided, we poll the api until the cluster is ready, or a timeout
	if r.args.createClusterWaitDuration > 0 {
		return waitForCluster(r.kotsAPI, cl.ID, r.args.createClusterWaitDuration)
	}

	return cl, nil
}

func generateClusterName() string {
	return namesgenerator.GetRandomName(0)
}

func waitForCluster(kotsRestClient *kotsclient.VendorV3Client, id string, duration time.Duration) (*types.Cluster, error) {
	start := time.Now()
	for {
		cluster, err := kotsRestClient.GetCluster(id)
		if err != nil {
			return nil, errors.Wrap(err, "get cluster")
		}

		if cluster.Status == types.ClusterStatusRunning {
			return cluster, nil
		} else if cluster.Status == types.ClusterStatusError {
			return nil, errors.New("cluster failed to provision")
		} else {
			if time.Now().After(start.Add(duration)) {
				return cluster, nil
			}
		}

		time.Sleep(time.Second * 5)
	}
}
