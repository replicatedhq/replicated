package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMGetSSHEndpoint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-endpoint [VM_ID]",
		Short: "Get the SSH endpoint of a VM",
		Long: `Get the SSH endpoint and port of a VM.

If a VM ID is provided, it will directly get the endpoint for that VM. Otherwise, if multiple VMs are available, you will be prompted to select the VM you want to get the endpoint for.

The output will be in the format: hostname:port`,
		Example: `# Get SSH endpoint for a specific VM by ID
replicated vm get ssh-endpoint <id>

# Get SSH endpoint (interactive selection if multiple VMs exist)
replicated vm get ssh-endpoint`,
		RunE:              r.getVMSSHEndpoint,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) getVMSSHEndpoint(cmd *cobra.Command, args []string) error {
	var vmID string
	if len(args) > 0 {
		vmID = args[0]
	}

	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return errors.Wrap(err, "list vms")
	}

	var vm *types.VM
	if vmID != "" {
		for _, v := range vms {
			if v.ID == vmID {
				vm = v
				break
			}
		}
		if vm == nil {
			return errors.Errorf("VM %s not found", vmID)
		}
	} else {
		runningVMs := filterVMsByStatus(vms, types.VMStatusRunning)
		if len(runningVMs) == 0 {
			return handleNoRunningVMs(vms)
		}
		selectedVM, err := selectVM(runningVMs, "Select a VM to get SSH endpoint for")
		if err != nil {
			return err
		}
		vm = selectedVM
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", vm.ID)
	}

	fmt.Printf("ssh://%s:%d\n", vm.DirectSSHEndpoint, vm.DirectSSHPort)
	return nil
}
