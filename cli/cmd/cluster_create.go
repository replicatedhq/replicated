package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
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

	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "kind", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "v1.25.3", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "node", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterVCpus, "vcpus", int64(4), "vCPUs to request per node")
	cmd.Flags().StringVar(&r.args.createClusterVCpuType, "vcpu-type", "", "vCPU type to request per node")
	cmd.Flags().Int64Var(&r.args.createClusterMemoryMiB, "memory", int64(4096), "Memory (MiB) to request per node (Default: 4096)")
	cmd.Flags().Int64Var(&r.args.createClusterDiskMiB, "disk", int64(51200), "Disk Size (MiB) to request per node (Default: 51200)")
	cmd.Flags().StringVar(&r.args.createClusterDiskType, "disk-type", "", "Disk type to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "2h", "Cluster TTL (duration, max 48h)")
	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")
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
		VCpuType:               r.args.createClusterVCpuType,
		MemoryMiB:              r.args.createClusterMemoryMiB,
		DiskMiB:                r.args.createClusterDiskMiB,
		DiskType:               r.args.createClusterDiskType,
		TTL:                    r.args.createClusterTTL,
		DryRun:                 r.args.createClusterDryRun,
	}
	cl, ve, err := kotsRestClient.CreateCluster(opts)
	if errors.Cause(err) == kotsclient.ErrForbidden {
		return errors.New("This command is not available for your account or team. Please contact your customer success representative for more information.")
	}
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	if ve != nil && len(ve.Errors) > 0 {
		if len(ve.SupportedDistributions) > 0 {
			return fmt.Errorf("%s\n\nSupported Kubernetes distributions and versions are:\n%s", errors.New(strings.Join(ve.Errors, ",")), supportedDistributions(ve.SupportedDistributions))
		} else {
			return fmt.Errorf("%s", errors.New(strings.Join(ve.Errors, ",")))
		}

	}
	if opts.DryRun {
		_, err := fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}
	return print.Cluster(r.outputFormat, r.w, cl)
}

func supportedDistributions(supportedDistributions map[string][]string) string {
	var supported []string
	for k, vv := range supportedDistributions {
		// assume that the vv is semver and sort
		vs := make([]*semver.Version, len(vv))
		for i, r := range vv {
			v, err := semver.NewVersion(r)
			if err != nil {
				// just don't include it
				continue
			}

			vs[i] = v
		}

		sort.Sort(semver.Collection(vs))

		supported = append(supported, fmt.Sprintf("  %s:", k))
		for _, v := range vs {
			supported = append(supported, fmt.Sprintf("    %s", v.Original()))
		}
	}
	return strings.Join(supported, "\n")
}
