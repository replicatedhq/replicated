package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "node-count", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterVCpus, "vcpus", int64(4), "vCPUs to request per node")
	cmd.Flags().Int64Var(&r.args.createClusterMemoryGiB, "memory", int64(4), "Memory (GiB) to request per node")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node (Default: 50)")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "2h", "Cluster TTL (duration, max 48h)")
	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for cluster to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "the type of instance to use for cloud-based clusters (e.g. x5.xlarge)")

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
		MemoryGiB:              r.args.createClusterMemoryGiB,
		DiskGiB:                r.args.createClusterDiskGiB,
		TTL:                    r.args.createClusterTTL,
		DryRun:                 r.args.createClusterDryRun,
		InstanceType:           r.args.createClusterInstanceType,
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

	// if the wait flag was provided, we poll the api until the cluster is ready, or a timeout
	if r.args.createClusterWaitDuration > 0 {
		return r.waitForCluster(cl.ID, r.args.createClusterWaitDuration)
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

func (r *runners) waitForCluster(id string, duration time.Duration) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	start := time.Now()
	for {
		clusters, err := kotsRestClient.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}

		for _, cluster := range clusters {
			if cluster.ID == id {
				if cluster.Status == "running" {
					return print.Cluster(r.outputFormat, r.w, cluster)
				} else if cluster.Status == "error" {
					return errors.New("cluster failed to provision")
				} else {
					if time.Now().After(start.Add(duration)) {
						return print.Cluster(r.outputFormat, r.w, cluster)
					}
				}
			}
		}

		time.Sleep(time.Second * 5)
	}
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
