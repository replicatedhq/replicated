package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
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

	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "kind", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "v1.25.3", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "node", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterVCpus, "vcpus", int64(4), "vCPUs to request per node")
	cmd.Flags().Int64Var(&r.args.createClusterMemoryMiB, "memory", int64(4096), "Memory (MiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "1h", "Cluster TTL (duration)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		NodeCount:              r.args.createClusterNodeCount,
		VCpus:                  r.args.createClusterVCpus,
		MemoryMiB:              r.args.createClusterMemoryMiB,
		TTL:                    r.args.createClusterTTL,
	}
	cl, err := kotsRestClient.CreateCluster(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}
