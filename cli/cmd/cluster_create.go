package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create test clusters",
		Long:         `create test clusters`,
		RunE:         r.createCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "cluster name")
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringVar(&r.args.createClusterOSDistribution, "os-distribution", "ubuntu", "OS distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterOSVersion, "os-version", "jammy", "OS version of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "kubernetes-distribution", "kind", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "kubernetes-version", "latest", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "node-count", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterVCpus, "vcpus", int64(4), "vCPUs to request per node")
	cmd.Flags().Int64Var(&r.args.createClusterMemoryMiB, "memory-mib", int64(4096), "Memory (MiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "1h", "Cluster TTL (duration)")

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		OSDistribution:         r.args.createClusterOSDistribution,
		OSVersion:              r.args.createClusterOSVersion,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		NodeCount:              r.args.createClusterNodeCount,
		VCpus:                  r.args.createClusterVCpus,
		MemoryMiB:              r.args.createClusterMemoryMiB,
		TTL:                    r.args.createClusterTTL,
	}
	_, err := kotsRestClient.CreateCluster(opts)
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}
