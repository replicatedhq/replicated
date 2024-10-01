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

var ErrVMWaitDurationExceeded = errors.New("wait duration exceeded")

func (r *runners) InitVMCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create test VMs",
		Long:         `Create test VMs.`,
		SilenceUsage: true,
		RunE:         r.createVM,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "VM name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesDistribution, "distribution", "", "Distribution of the vm to provision")
	cmd.Flags().StringVar(&r.args.createClusterKubernetesVersion, "version", "", "Vversion to provision (format is distribution dependent)")
	cmd.Flags().StringVar(&r.args.createClusterIPFamily, "ip-family", "", "IP Family to use for the vm (ipv4|ipv6|dual).")
	cmd.Flags().IntVar(&r.args.createClusterNodeCount, "nodes", int(1), "Node count")
	cmd.Flags().Int64Var(&r.args.createClusterDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "VM TTL (duration, max 48h)")
	cmd.Flags().DurationVar(&r.args.createClusterWaitDuration, "wait", time.Second*0, "Wait duration for VM to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createClusterInstanceType, "instance-type", "", "The type of instance to use (e.g. r1.medium)")
	cmd.Flags().StringArrayVar(&r.args.createClusterNodeGroups, "nodegroup", []string{}, "Node group to create (name=?,instance-type=?,nodes=?,disk=? format, can be specified multiple times). For each nodegroup, one of the following flags must be specified: name, instance-type, nodes or disk.")

	cmd.Flags().StringArrayVar(&r.args.createClusterTags, "tag", []string{}, "Tag to apply to the VM (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.createClusterDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	_ = cmd.MarkFlagRequired("distribution")

	return cmd
}

func (r *runners) createVM(_ *cobra.Command, args []string) error {
	if r.args.createClusterName == "" {
		r.args.createClusterName = namesgenerator.GetRandomName(0)
	}

	tags, err := parseTags(r.args.createClusterTags)
	if err != nil {
		return errors.Wrap(err, "parse tags")
	}

	nodeGroups, err := parseVMNodeGroups(r.args.createClusterNodeGroups)
	if err != nil {
		return errors.Wrap(err, "parse node groups")
	}

	opts := kotsclient.CreateVMOpts{
		Name:         r.args.createClusterName,
		Distribution: r.args.createClusterKubernetesDistribution,
		Version:      r.args.createClusterKubernetesVersion,
		IPFamily:     r.args.createClusterIPFamily,
		NodeCount:    r.args.createClusterNodeCount,
		DiskGiB:      r.args.createClusterDiskGiB,
		TTL:          r.args.createClusterTTL,
		InstanceType: r.args.createClusterInstanceType,
		NodeGroups:   nodeGroups,
		Tags:         tags,
		DryRun:       r.args.createClusterDryRun,
	}

	vm, err := r.createAndWaitForVM(opts)
	if err != nil {
		if errors.Cause(err) == ErrVMWaitDurationExceeded {
			defer func() {
				os.Exit(124)
			}()
		} else {
			return err
		}
	}

	if opts.DryRun {
		estimatedCostMessage := fmt.Sprintf("Estimated cost: %s (if run to TTL of %s)", print.CreditsToDollarsDisplay(vm.EstimatedCost), vm.TTL)
		_, err := fmt.Fprintln(r.w, estimatedCostMessage)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.VM(r.outputFormat, r.w, vm)
}

func (r *runners) createAndWaitForVM(opts kotsclient.CreateVMOpts) (*types.VM, error) {
	vm, ve, err := r.kotsAPI.CreateVM(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create vm")
	}

	if ve != nil && ve.Message != "" {
		if ve.ValidationError != nil && len(ve.ValidationError.Errors) > 0 {
			if len(ve.ValidationError.SupportedDistributions) > 0 {
				_ = print.VMVersions("table", r.w, ve.ValidationError.SupportedDistributions)
			}
		}
		return nil, errors.New(ve.Message)
	}

	if opts.DryRun {
		return vm, nil
	}

	// if the wait flag was provided, we poll the api until the vm is ready, or a timeout
	if r.args.createClusterWaitDuration > 0 {
		return waitForVM(r.kotsAPI, vm.ID, r.args.createClusterWaitDuration)
	}

	return vm, nil
}

func waitForVM(kotsRestClient *kotsclient.VendorV3Client, id string, duration time.Duration) (*types.VM, error) {
	start := time.Now()
	for {
		vm, err := kotsRestClient.GetVM(id)
		if err != nil {
			return nil, errors.Wrap(err, "get vm")
		}

		if vm.Status == types.ClusterStatusRunning {
			return vm, nil
		} else if vm.Status == types.ClusterStatusError || vm.Status == types.ClusterStatusUpgradeError {
			return nil, errors.New("vm failed to provision")
		} else {
			if time.Now().After(start.Add(duration)) {
				// In case of timeout, return the vm and a WaitDurationExceeded error
				return vm, ErrWaitDurationExceeded
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func parseVMNodeGroups(nodeGroups []string) ([]kotsclient.VMNodeGroup, error) {
	parsedNodeGroups := []kotsclient.VMNodeGroup{}
	for _, nodeGroup := range nodeGroups {
		field := strings.Split(nodeGroup, ",")
		ng := kotsclient.VMNodeGroup{}
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

		parsedNodeGroups = append(parsedNodeGroups, ng)
	}
	return parsedNodeGroups, nil
}
