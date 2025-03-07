package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSSH(parent *cobra.Command) *cobra.Command {
	var sshUser string

	cmd := &cobra.Command{
		Use:   "ssh [VM_ID]",
		Short: "SSH into a VM",
		Long: `Connect to a VM using SSH.

If a VM ID is provided, it will directly connect to that VM. Otherwise, if multiple VMs are available, you will be prompted to select the VM you want to connect to. The command will then establish an SSH connection to the selected VM using the appropriate credentials and configuration.

The SSH user can be specified in order of precedence:
1. By specifying the -u flag
2. REPLICATED_SSH_USER environment variable
3. GITHUB_ACTOR environment variable (from GitHub Actions)
4. GITHUB_USER environment variable

Note: Only running VMs can be connected to via SSH.`,
		Example: `# SSH into a specific VM by ID
replicated vm ssh <id>

# SSH into a VM with a specific user
replicated vm ssh <id> -u myuser

# SSH into a VM (interactive selection if multiple VMs exist)
replicated vm ssh`,
		RunE:              r.sshVM,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&sshUser, "user", "u", "", "SSH user to connect with")

	return cmd
}

// isVMRunning checks if a VM is running
func isVMRunning(vm *types.VM) bool {
	return vm.Status == types.VMStatusRunning
}

func (r *runners) sshVM(cmd *cobra.Command, args []string) error {
	if err := r.initVMClient(); err != nil {
		return err
	}

	sshUser, _ := cmd.Flags().GetString("user")

	// Get VM ID - either directly provided or selected
	var vmID string

	if len(args) == 1 {
		// VM ID provided directly
		vmID = args[0]

		// Check VM status before connecting
		vm, err := r.kotsAPI.GetVM(vmID)
		if err != nil {
			return errors.Wrap(err, "failed to get VM")
		}

		// Only connect if VM is running
		if !isVMRunning(vm) {
			return fmt.Errorf("VM %s is not running (current status: %s). Cannot connect to a non-running VM", vmID, vm.Status)
		}
	} else {
		// Need to select a VM
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "failed to list VMs")
		}

		// Filter to only show running VMs
		runningVMs := filterVMsByStatus(vms, types.VMStatusRunning)
		if len(runningVMs) == 0 {
			return handleNoRunningVMs(vms)
		}

		// Select VM from running VMs
		selectedVM, err := selectVM(runningVMs, "Select VM to SSH into")
		if err != nil {
			return errors.Wrap(err, "failed to select VM")
		}

		vmID = selectedVM.ID
	}

	return r.kotsAPI.SSHIntoVM(vmID, sshUser)
}

// handleNoRunningVMs handles the case when no running VMs are found
// It displays information about non-running VMs if they exist
func handleNoRunningVMs(vms []*types.VM) error {
	// Show information about non-running VMs if they exist
	nonTerminatedVMs := filterVMsByStatus(vms, "")
	if len(nonTerminatedVMs) == 0 {
		return errors.New("no active VMs found")
	}

	// List non-running VMs with their statuses
	fmt.Println("No running VMs found. The following VMs are available but not running:")
	for _, vm := range nonTerminatedVMs {
		if vm.Status != types.VMStatusTerminated {
			fmt.Printf("  • %s (ID: %s, Status: %s)\n", vm.Name, vm.ID, vm.Status)
		}
	}
	return errors.New("SSH connection requires a running VM. Please start a VM before connecting")
}

// filterVMsByStatus returns a slice of VMs with the specified status
// If status is empty, returns all VMs
func filterVMsByStatus(vms []*types.VM, status types.VMStatus) []*types.VM {
	var filteredVMs []*types.VM
	for _, vm := range vms {
		if status == "" || vm.Status == status {
			// If status is empty, include all VMs except terminated ones
			if status == "" && vm.Status == types.VMStatusTerminated {
				continue
			}
			filteredVMs = append(filteredVMs, vm)
		}
	}
	return filteredVMs
}

// selectVM prompts the user to select a VM from the list
func selectVM(vms []*types.VM, label string) (*types.VM, error) {
	var vmOptions []string
	for _, vm := range vms {
		vmOptions = append(vmOptions, fmt.Sprintf("%s (%s)", vm.Name, vm.Status))
	}

	prompt := promptui.Select{
		Label: label,
		Items: vmOptions,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return vms[index], nil
}
