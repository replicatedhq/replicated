package cmd

import (
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
1. -u flag
2. REPLICATED_SSH_USER environment variable
3. GITHUB_ACTOR environment variable (from GitHub Actions)
4. GITHUB_USER environment variable`,
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

func (r *runners) sshVM(cmd *cobra.Command, args []string) error {
	if err := r.initVMClient(); err != nil {
		return err
	}

	sshUser, _ := cmd.Flags().GetString("user")

	// If VM ID is provided, directly SSH into it
	if len(args) == 1 {
		return r.kotsAPI.SSHIntoVM(args[0], sshUser)
	}

	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to list VMs")
	}

	// Filter out terminated VMs
	activeVMs := filterActiveVMs(vms)
	if len(activeVMs) == 0 {
		return errors.New("no active VMs found")
	}

	// If only one VM, use it directly
	if len(activeVMs) == 1 {
		return r.kotsAPI.SSHIntoVM(activeVMs[0].ID, sshUser)
	}

	// Create VM selection prompt
	selectedVM, err := selectVM(activeVMs)
	if err != nil {
		return errors.Wrap(err, "failed to select VM")
	}

	return r.kotsAPI.SSHIntoVM(selectedVM.ID, sshUser)
}

// filterActiveVMs returns a slice of non-terminated VMs
func filterActiveVMs(vms []*types.VM) []*types.VM {
	var activeVMs []*types.VM
	for _, vm := range vms {
		if vm.Status != types.VMStatusTerminated {
			activeVMs = append(activeVMs, vm)
		}
	}
	return activeVMs
}

// selectVM prompts the user to select a VM from the list
func selectVM(vms []*types.VM) (*types.VM, error) {
	var vmOptions []string
	for _, vm := range vms {
		vmOptions = append(vmOptions, vm.Name)
	}

	prompt := promptui.Select{
		Label: "Select VM to connect to",
		Items: vmOptions,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return vms[index], nil
}
