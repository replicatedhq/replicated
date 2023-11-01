package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create a node group",
		Long:         ``,
		RunE:         r.createClusterNodeGroup,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterNodeGroupInstanceType, "instance-type", "", "Instance type for the new node group")
	cmd.Flags().IntVar(&r.args.createClusterNodeGroupNodeCount, "nodes", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterNodeGroupDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().DurationVar(&r.args.createClusterNodeGroupWaitDuration, "wait", time.Second*0, "Wait duration for node group to be ready (leave empty to not wait)")

	cmd.MarkFlagRequired("cluster-id")

	return cmd
}

func (r *runners) createClusterNodeGroup(cmd *cobra.Command, args []string) error {
	if r.args.createClusterName == "" {
		r.args.createClusterName = util.GenerateName()
	}

	opts := kotsclient.CreateClusterNodeGroupOpts{
		ClusterID:    args[0],
		Name:         r.args.createClusterNodeGroupName,
		NodeCount:    r.args.createClusterNodeGroupNodeCount,
		DiskGiB:      r.args.createClusterNodeGroupDiskGiB,
		InstanceType: r.args.createClusterNodeGroupInstanceType,
	}
	ng, err := r.createAndWaitForNodeGroup(opts)
	if err != nil {
		return err
	}

	return print.ClusterNodeGroup(r.outputFormat, r.w, ng)
}

func (r *runners) createAndWaitForNodeGroup(opts kotsclient.CreateClusterNodeGroupOpts) (*types.ClusterNodeGroup, error) {
	ng, ve, err := r.kotsAPI.CreateClusterNodeGroup(opts.ClusterID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster node group")
	}

	if ve != nil && ve.Message != "" {
		if ve.ValidationError != nil && len(ve.ValidationError.Errors) > 0 {
			if len(ve.ValidationError.SupportedInstanceTypes) > 0 {
				_ = print.ClusterInstanceTypes("table", r.w, ve.ValidationError.SupportedInstanceTypes)
			}
		}
		return nil, errors.New(ve.Message)
	}

	if r.args.createClusterNodeGroupWaitDuration > 0 {
		return waitForNodeGroup(r.kotsAPI, ng.ID, r.args.createClusterNodeGroupWaitDuration)
	}

	return ng, nil
}

func waitForNodeGroup(kotsRestClient *kotsclient.VendorV3Client, id string, duration time.Duration) (*types.ClusterNodeGroup, error) {
	start := time.Now()
	for {
		ng, err := kotsRestClient.GetClusterNodeGroup(id)
		if err != nil {
			return nil, errors.Wrap(err, "get cluster node group")
		}

		if time.Now().After(start.Add(duration)) {
			return ng, nil
		}
		time.Sleep(time.Second * 5)
	}
}
