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
		Short: "Create test clusters",
		Long: `Create test clusters.

This is a beta feature, with some known limitations:
https://docs.replicated.com/vendor/testing-how-to#limitations`,
		SilenceUsage: true,
		RunE:         r.createCluster,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "Cluster name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "", "Kubernetes distribution of the cluster to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "", "Kubernetes version to provision (format is distribution dependent)")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "nodes", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "Cluster TTL (duration, max 48h)")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for cluster to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "The type of instance to use (e.g. m6i.large)")
	cmd.Flags().StringArrayVar(&r.args.createClusterNodeGroups, "nodegroup", []string{}, "Node group to create (name=?,instance-type=?,nodes=?,disk=? format, can be specified multiple times)")

	cmd.Flags().StringArrayVar(&r.args.createClusterTags, "tag", []string{}, "Tag to apply to the cluster (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

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

	nodeGroups, err := parseNodeGroups(r.args.createClusterNodeGroups)
	if err != nil {
		return errors.Wrap(err, "parse node groups")
	}

	opts := kotsclient.CreateClusterOpts{
		Name:                   r.args.createClusterName,
		KubernetesDistribution: r.args.createClusterKubernetesDistribution,
		KubernetesVersion:      r.args.createClusterKubernetesVersion,
		NodeCount:              r.args.createClusterNodeCount,
		DiskGiB:                r.args.createClusterDiskGiB,
		TTL:                    r.args.createClusterTTL,
		InstanceType:           r.args.createClusterInstanceType,
		NodeGroups:             nodeGroups,
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
