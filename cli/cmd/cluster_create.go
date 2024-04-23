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
		Use:          "create",
		Short:        "Create test clusters",
		Long:         `Create test clusters.`,
		SilenceUsage: true,
		RunE:         r.createCluster,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "Cluster name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().StringVar(&r.args.createClusterLicenseID, "license-id", "", "License ID to use for the installation (required for Embedded Cluster distribution)")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "Cluster TTL (duration, max 48h)")
	cmd.Flags().StringArrayVar(&r.args.createClusterTags, "tag", []string{}, "Tag to apply to the cluster (key=value format, can be specified multiple times)")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for cluster to be ready (leave empty to not wait)")

	// the CLI supports setting the default node group via separate flags or the --default-nodegroup flag
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "nodes", int(1), "Node count")
	cmd.Flags().StringVar(&r.args.createClusterMinNodeCount, "min-nodes", 0, "Minimum Node count (non-negative number) (only for EKS, AKS and GKE clusters).")
	cmd.Flags().StringVar(&r.args.createClusterMaxNodeCount, "max-nodes", 0, "Maximum Node count (non-negative number) (only for EKS, AKS and GKE clusters).")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "The type of instance to use (e.g. m6i.large)")

	cmd.Flags().StringVar(&r.args.createClusterDefaultNodeGroup, "default-nodegroup", "", "Node group to create (name=?,instance-type=?,nodes=?,min-nodes=?,max-nodes=?,disk=? format)")

	// the CLI supports creating additional node groups (not default) via the --additional-nodegroup flag
	cmd.Flags().StringArrayVar(&r.args.createClusterAdditionalNodeGroups, "additional-nodegroup", []string{}, "Node group to create (name=?,instance-type=?,nodes=?,min-nodes=?,max-nodes=?,disk=? format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	_ = cmd.MarkFlagRequired("distribution")

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	if r.args.createClusterName == "" {
		r.args.createClusterName = generateClusterName()
	}

	tags, err := parseTags(r.args.createClusterTags)
	if err != nil {
		return errors.Wrap(err, "parse tags")
	}

	additionalNodeGroups, err := parseNodeGroups(r.args.createClusterAdditionalNodeGroups)
	if err != nil {
		return errors.Wrap(err, "parse node groups")
	}

	defaultNodeGroup := kotsclient.NodeGroup{}
	if r.args.createClusterDefaultNodeGroup != "" {
		parsed, err := parseNodeGroups([]string{r.args.createClusterDefaultNodeGroup})
		if err != nil {
			return errors.Wrap(err, "parse default node group")
		}
		if len(parsed) != 1 {
			return errors.New("invalid default node group format")
		}
		defaultNodeGroup = parsed[0]
	} else {
		defaultNodeGroup = kotsclient.NodeGroup{
			Name:         "default",
			InstanceType: r.args.createClusterInstanceType,
			Disk:         int(r.args.createClusterDiskGiB),
		}

		if r.args.createClusterNodeCount > 0 {
			defaultNodeGroup.Nodes = r.args.createClusterNodeCount
		} else {
			defaultNodeGroup.MinNodes = &r.args.createClusterMinNodeCount
			defaultNodeGroup.MaxNodes = &r.args.createClusterMaxNodeCount
		}
	}

	// the default node group is always set in defaultNodeGroup, regardless of
	// how the user passed the params in

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		LicenseID:              r.args.createClusterLicenseID,
		TTL:                    r.args.createClusterTTL,
		InstanceType:           r.args.createClusterInstanceType,
		AdditionalNodeGroups:   additionalNodeGroups,
		DefaultNodeGroup:       defaultNodeGroup,
		Tags:                   tags,
		DryRun:                 r.args.createClusterDryRun,
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

	if ve != nil && ve.Message != "" {
		if ve.ValidationError != nil && len(ve.ValidationError.Errors) > 0 {
			if len(ve.ValidationError.SupportedDistributions) > 0 {
				_ = print.ClusterVersions("table", r.w, ve.ValidationError.SupportedDistributions)
			}
		}
		return nil, errors.New(ve.Message)
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

func parseNodeGroups(nodeGroups []string) ([]kotsclient.NodeGroup, error) {
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

		// check if instanceType, nodes and disk are set (required)
		if ng.InstanceType == "" || ng.Nodes == 0 || ng.Disk == 0 {
			return nil, errors.Errorf("invalid node group format: %s", nodeGroup)
		}
		parsedNodeGroups = append(parsedNodeGroups, ng)
	}
	return parsedNodeGroups, nil
}
