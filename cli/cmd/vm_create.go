package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
	"github.com/spf13/cobra"
)

var ErrVMWaitDurationExceeded = errors.New("wait duration exceeded")

func (r *runners) InitVMCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create one or more test VMs with specified distribution, version, and configuration options.",
		Long: `Create one or more test VMs with a specified distribution, version, and a variety of customizable configuration options.

This command allows you to provision VMs with different distributions (e.g., Ubuntu, RHEL), versions, instance types, and more. You can set the number of VMs to create, disk size, and specify the network to use. If no network is provided, a new network will be created automatically. You can also assign tags to your VMs and use a TTL (Time-To-Live) to define how long the VMs should live.

By default, the command provisions one VM, but you can customize the number of VMs to create by using the "--count" flag. Additionally, you can use the "--dry-run" flag to simulate the creation without actually provisioning the VMs.

The command also supports a "--wait" flag to wait for the VMs to be ready before returning control, with a customizable timeout duration.`,
		Example: `# Create a single Ubuntu 20.04 VM
replicated vm create --distribution ubuntu --version 20.04

# Create 3 Ubuntu 22.04 VMs
replicated vm create --distribution ubuntu --version 22.04 --count 3

# Create 5 Ubuntu VMs with a custom instance type and disk size
replicated vm create --distribution ubuntu --version 20.04 --count 5 --instance-type r1.medium --disk 100

# Create a VM with an SSH public key
replicated vm create --distribution ubuntu --version 20.04 --ssh-public-key ~/.ssh/id_rsa.pub

# Create a VM with multiple SSH public keys
replicated vm create --distribution ubuntu --version 20.04 --ssh-public-key ~/.ssh/id_rsa.pub --ssh-public-key ~/.ssh/id_ed25519.pub`,
		SilenceUsage: true,
		RunE:         r.createVM,
		Args:         cobra.NoArgs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createVMName, "name", "", "VM name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createVMDistribution, "distribution", "", "Distribution of the VM to provision")
	cmd.RegisterFlagCompletionFunc("distribution", r.completeVMDistributions)
	cmd.Flags().StringVar(&r.args.createVMVersion, "version", "", "Version to provision (format is distribution dependent)")
	cmd.RegisterFlagCompletionFunc("version", r.completeVMVersions)
	cmd.Flags().IntVar(&r.args.createVMCount, "count", int(1), "Number of matching VMs to create")
	cmd.Flags().Int64Var(&r.args.createVMDiskGiB, "disk", int64(50), "Disk Size (GiB) to request per node")
	cmd.Flags().StringVar(&r.args.createVMTTL, "ttl", "", "VM TTL (duration, max 48h)")
	cmd.Flags().DurationVar(&r.args.createVMWaitDuration, "wait", time.Second*0, "Wait duration for VM(s) to be ready (leave empty to not wait)")
	cmd.Flags().StringVar(&r.args.createVMInstanceType, "instance-type", "", "The type of instance to use (e.g. r1.medium)")
	cmd.RegisterFlagCompletionFunc("instance-type", r.completeVMInstanceTypes)
	cmd.Flags().StringVar(&r.args.createVMNetwork, "network", "", "The network to use for the VM(s). If not supplied, create a new network")

	cmd.Flags().StringArrayVar(&r.args.createVMTags, "tag", []string{}, "Tag to apply to the VM (key=value format, can be specified multiple times)")
	cmd.Flags().StringArrayVar(&r.args.createVMPublicKeys, "ssh-public-key", []string{}, "Path to SSH public key file to add to the VM (can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.createVMDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	_ = cmd.MarkFlagRequired("distribution")

	return cmd
}

func (r *runners) createVM(_ *cobra.Command, args []string) error {
	if r.args.createVMName == "" {
		r.args.createVMName = namesgenerator.GetRandomName(0)
	}

	tags, err := parseTags(r.args.createVMTags)
	if err != nil {
		return errors.Wrap(err, "parse tags")
	}

	var publicKeys []string
	for _, keyPath := range r.args.createVMPublicKeys {
		publicKey, err := util.ReadAndValidatePublicKey(keyPath)
		if err != nil {
			return errors.Wrap(err, "validate public key")
		}
		publicKeys = append(publicKeys, publicKey)
	}

	opts := kotsclient.CreateVMOpts{
		Name:         r.args.createVMName,
		Distribution: r.args.createVMDistribution,
		Version:      r.args.createVMVersion,
		Count:        r.args.createVMCount,
		DiskGiB:      r.args.createVMDiskGiB,
		Network:      r.args.createVMNetwork,
		TTL:          r.args.createVMTTL,
		InstanceType: r.args.createVMInstanceType,
		Tags:         tags,
		PublicKeys:   publicKeys,
		DryRun:       r.args.createVMDryRun,
	}

	vms, err := r.createAndWaitForVM(opts)
	if err != nil {
		if errors.Cause(err) == ErrVMWaitDurationExceeded {
			defer os.Exit(124)
		} else {
			return err
		}
	}

	if opts.DryRun {
		// This should not happen, as count should be > 0
		if len(vms) == 0 {
			return errors.New("no vm will be created")
		}
		estimatedCostMessage := fmt.Sprintf("Estimated cost: %s (if run to TTL of %s)", print.CreditsToDollarsDisplay(vms[0].EstimatedCost), vms[0].TTL)
		_, err := fmt.Fprintln(r.w, estimatedCostMessage)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.VMs(r.outputFormat, r.w, vms, true)
}

func (r *runners) createAndWaitForVM(opts kotsclient.CreateVMOpts) ([]*types.VM, error) {
	vms, ve, err := r.kotsAPI.CreateVM(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create vm")
	}

	if ve != nil && ve.Message != "" {
		if ve.ValidationError != nil && len(ve.ValidationError.Errors) > 0 {
			if len(ve.ValidationError.SupportedDistributions) > 0 {
				if err := print.VMVersions("table", r.w, ve.ValidationError.SupportedDistributions); err != nil {
					return nil, errors.Wrap(err, "print vm versions")
				}
			}
		}
		return nil, errors.New(ve.Message)
	}

	if opts.DryRun {
		return vms, nil
	}

	// if the wait flag was provided, we poll the api until the vm is ready, or a timeout
	if r.args.createVMWaitDuration > 0 {
		return waitForVMs(r.kotsAPI, vms, r.args.createVMWaitDuration)
	}

	return vms, nil
}

func waitForVMs(kotsRestClient *kotsclient.VendorV3Client, vms []*types.VM, duration time.Duration) ([]*types.VM, error) {
	start := time.Now()
	runningVMs := map[string]*types.VM{}
	for {
		for _, vm := range vms {
			v, err := kotsRestClient.GetVM(vm.ID)
			if err != nil {
				return nil, errors.Wrap(err, "get vm")
			}

			if v.Status == types.VMStatus(types.VMStatusRunning) {
				runningVMs[v.ID] = v
				if len(runningVMs) == len(vms) {
					return mapToSlice(runningVMs), nil
				}
			} else if vm.Status == types.VMStatus(types.VMStatusError) {
				return nil, errors.New("vm failed to provision")
			} else {
				if time.Now().After(start.Add(duration)) {
					// In case of timeout, return the vm and a WaitDurationExceeded error
					return mapToSlice(runningVMs), ErrWaitDurationExceeded
				}
			}
		}

		time.Sleep(time.Second * 5)
	}
}

// Convert map of VMs to slice of VMs
func mapToSlice(vms map[string]*types.VM) []*types.VM {
	var slice []*types.VM
	for _, v := range vms {
		slice = append(slice, v)
	}
	return slice
}
