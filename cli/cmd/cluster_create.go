package cmd

import (
	"fmt"
	"os"
	"strconv"
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

var ErrWaitDurationExceeded = errors.New("wait duration exceeded")

func (r *runners) InitClusterCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create test clusters.",
		Long: `The 'cluster create' command provisions a new test cluster with the specified Kubernetes distribution and configuration. You can customize the cluster's size, version, node groups, disk space, IP family, and other parameters.

This command supports creating clusters on multiple Kubernetes distributions, including setting up node groups with different instance types and counts. You can also specify a TTL (Time-To-Live) to automatically terminate the cluster after a set duration.

Use the '--dry-run' flag to simulate the creation process and get an estimated cost without actually provisioning the cluster.`,
		Example: `  # Create a new cluster with basic configuration
  replicated cluster create --distribution eks --version 1.21 --nodes 3 --instance-type t3.large --disk 100 --ttl 24h

  # Create a cluster with a custom node group
  replicated cluster create --distribution eks --version 1.21 --nodegroup name=workers,instance-type=t3.large,nodes=5 --ttl 24h

  # Simulate cluster creation (dry-run)
  replicated cluster create --distribution eks --version 1.21 --nodes 3 --disk 100 --ttl 24h --dry-run

  # Create a cluster with autoscaling configuration
  replicated cluster create --distribution eks --version 1.21 --min-nodes 2 --max-nodes 5 --instance-type t3.large --ttl 24h

  # Create a cluster with multiple node groups
  replicated cluster create --distribution eks --version 1.21 \
    --nodegroup name=workers,instance-type=t3.large,nodes=3 \
    --nodegroup name=cpu-intensive,instance-type=c5.2xlarge,nodes=2 \
    --ttl 24h

  # Create a cluster with custom tags
  replicated cluster create --distribution eks --version 1.21 --nodes 3 --tag env=test --tag project=demo --ttl 24h

  # Create a cluster with addons
  replicated cluster create --distribution eks --version 1.21 --nodes 3 --addon object-store --ttl 24h`,
		SilenceUsage: true,
		RunE:         r.createCluster,
		Args:         cobra.NoArgs,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "Cluster name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().StringVar(&r.args.createClusterIPFamily, "ip-family", "", "IP Family to use for the cluster (ipv4|ipv6|dual).")
	cmd.Flags().StringVar(&r.args.createClusterLicenseID, "license-id", "", "License ID to use for the installation (required for Embedded Cluster distribution)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "nodes", int(1), "Node count")
	cmd.Flags().StringVar(&r.args.createClusterMinNodeCount, "min-nodes", "", "Minimum Node count (non-negative number) (only for EKS, AKS and GKE clusters).")
	cmd.Flags().StringVar(&r.args.createClusterMaxNodeCount, "max-nodes", "", "Maximum Node count (non-negative number) (only for EKS, AKS and GKE clusters).")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "Cluster TTL (duration, max 48h)")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for cluster to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "The type of instance to use (e.g. m6i.large)")
	cmd.Flags().StringArrayVar(&r.args.createClusterNodeGroups, "nodegroup", []string{}, "Node group to create (name=?,instance-type=?,nodes=?,min-nodes=?,max-nodes=?,disk=? format, can be specified multiple times). For each nodegroup, at least one flag must be specified. The flags min-nodes and max-nodes are mutually dependent.")

	cmd.Flags().StringArrayVar(&r.args.createClusterTags, "tag", []string{}, "Tag to apply to the cluster (key=value format, can be specified multiple times)")

	cmd.Flags().StringArrayVar(&r.args.createClusterAddons, "addon", []string{}, "Addons to install on the cluster (can be specified multiple times)")
	cmd.Flags().StringVar(&r.args.clusterAddonCreateObjectStoreBucket, "bucket-prefix", "", "A prefix for the bucket name to be created (required by '--addon object-store')")

	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	_ = cmd.MarkFlagRequired("distribution")

	return cmd
}

func (r *runners) validateAddonArgs(cmd *cobra.Command) error {
	for _, addon := range r.args.createClusterAddons {
		if addon == "object-store" {
			if r.args.clusterAddonCreateObjectStoreBucket == "" {
				if err := cmd.Help(); err != nil {
					return err
				}
				return errors.New("bucket-prefix is required for object-store addon")
			}
		} else {
			return errors.Errorf("unknown addon: %s", addon)
		}
	}

	return nil
}

func (r *runners) createCluster(cmd *cobra.Command, args []string) error {
	if r.args.createClusterName == "" {
		r.args.createClusterName = generateClusterName()
	}

	tags, err := parseTags(r.args.createClusterTags)
	if err != nil {
		return errors.Wrap(err, "parse tags")
	}

	nodeGroups, err := parseClusterNodeGroups(r.args.createClusterNodeGroups)
	if err != nil {
		return errors.Wrap(err, "parse node groups")
	}

	if err := r.validateAddonArgs(cmd); err != nil {
		return errors.Wrap(err, "validate addon args")
	}

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		IPFamily:               r.args.createClusterIPFamily,
		LicenseID:              r.args.createClusterLicenseID,
		NodeCount:              r.args.createClusterNodeCount,
		DiskGiB:                r.args.createClusterDiskGiB,
		TTL:                    r.args.createClusterTTL,
		InstanceType:           r.args.createClusterInstanceType,
		NodeGroups:             nodeGroups,
		Tags:                   tags,
		DryRun:                 r.args.createClusterDryRun,
	}
	if r.args.createClusterMinNodeCount != "" {
		minNodes, err := strconv.Atoi(r.args.createClusterMinNodeCount)
		if err != nil {
			return errors.Wrapf(err, "failed to parse min-nodes value: %s", r.args.createClusterMinNodeCount)
		}
		if minNodes < 0 {
			return errors.Errorf("min-nodes must be a non-negative number: %s", r.args.createClusterMinNodeCount)
		}
		opts.MinNodeCount = &minNodes
	}
	if r.args.createClusterMaxNodeCount != "" {
		maxNodes, err := strconv.Atoi(r.args.createClusterMaxNodeCount)
		if err != nil {
			return errors.Wrapf(err, "failed to parse max-nodes value: %s", r.args.createClusterMaxNodeCount)
		}
		if maxNodes < 0 {
			return errors.Errorf("max-nodes must be a non-negative number: %s", r.args.createClusterMaxNodeCount)
		}
		opts.MaxNodeCount = &maxNodes
	}
	cl, err := r.createAndWaitForCluster(opts)
	if err != nil {
		if errors.Cause(err) == ErrWaitDurationExceeded {
			defer func() {
				os.Exit(124)
			}()
		} else {
			return err
		}
	}

	if r.args.createClusterAddons != nil {
		if err := r.createAndWaitForAddons(cl.ID); err != nil {
			return err
		}
	}

	if opts.DryRun {
		estimatedCostMessage := fmt.Sprintf("Estimated cost: %s (if run to TTL of %s)", print.CreditsToDollarsDisplay(cl.EstimatedCost), cl.TTL)
		_, err := fmt.Fprintln(r.w, estimatedCostMessage)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(r.w, "Dry run succeeded for cluster create.")
		return err
	}

	return print.Cluster(r.outputFormat, r.w, cl)
}

func (r *runners) createAndWaitForAddons(clusterID string) error {
	for _, addon := range r.args.createClusterAddons {
		if addon == "object-store" {
			// ClusterID, DryRun, Duration and Output are common to all
			// addons and cluster create, so they are inherited from the
			// cluster create command.
			r.args.clusterAddonCreateObjectStoreClusterID = clusterID
			r.args.clusterAddonCreateObjectStoreDryRun = r.args.createClusterDryRun
			r.args.clusterAddonCreateObjectStoreDuration = r.args.createClusterWaitDuration
			r.args.clusterAddonCreateObjectStoreOutput = r.outputFormat

			err := r.clusterAddonCreateObjectStoreCreateRun()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *runners) createAndWaitForCluster(opts kotsclient.CreateClusterOpts) (*types.Cluster, error) {
	cl, ve, err := r.kotsAPI.CreateCluster(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster")
	}

	if ve != nil && ve.Message != "" {
		if ve.ValidationError != nil && len(ve.ValidationError.Errors) > 0 {
			if len(ve.ValidationError.SupportedDistributions) > 0 {
				_ = print.ClusterVersions("table", r.w, ve.ValidationError.SupportedDistributions)
			}
		}
		return nil, errors.New(ve.Message)
	}

	if opts.DryRun {
		return cl, nil
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
		} else if cluster.Status == types.ClusterStatusError || cluster.Status == types.ClusterStatusUpgradeError {
			return nil, errors.New("cluster failed to provision")
		} else {
			if time.Now().After(start.Add(duration)) {
				// In case of timeout, return the cluster and a WaitDurationExceeded error
				return cluster, ErrWaitDurationExceeded
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func parseClusterNodeGroups(nodeGroups []string) ([]kotsclient.NodeGroup, error) {
	parsedNodeGroups := []kotsclient.NodeGroup{}
	for _, nodeGroup := range nodeGroups {
		field := strings.Split(nodeGroup, ",")
		ng := kotsclient.NodeGroup{}
		for _, f := range field {
			fieldParsed := strings.SplitN(f, "=", 2)
			if len(fieldParsed) != 2 {
				return nil, errors.Errorf("invalid node group format: %s", nodeGroup)
			}
			parsedFieldKey := fieldParsed[0]
			parsedFieldValue := fieldParsed[1]
			switch parsedFieldKey {
			case "name":
				ng.Name = parsedFieldValue
			case "instance-type":
				ng.InstanceType = parsedFieldValue
			case "nodes":
				nodes, err := strconv.Atoi(parsedFieldValue)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse nodes value: %s", parsedFieldValue)
				}
				ng.Nodes = nodes
			case "min-nodes":
				minNodes, err := strconv.Atoi(parsedFieldValue)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse min-nodes value: %s", parsedFieldValue)
				}
				if minNodes < 0 {
					return nil, errors.Errorf("min-nodes must be a non-negative number: %s", parsedFieldValue)
				}
				ng.MinNodes = &minNodes
			case "max-nodes":
				maxNodes, err := strconv.Atoi(parsedFieldValue)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse max-nodes value: %s", parsedFieldValue)
				}
				if maxNodes < 0 {
					return nil, errors.Errorf("max-nodes must be a non-negative number: %s", parsedFieldValue)
				}
				ng.MaxNodes = &maxNodes
			case "disk":
				diskSize, err := strconv.Atoi(parsedFieldValue)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse disk value: %s", parsedFieldValue)
				}
				ng.Disk = diskSize
			default:
				return nil, errors.Errorf("invalid node group field: %s", parsedFieldKey)
			}
		}

		parsedNodeGroups = append(parsedNodeGroups, ng)
	}
	return parsedNodeGroups, nil
}
