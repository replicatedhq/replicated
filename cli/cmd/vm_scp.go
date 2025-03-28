package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSCP(parent *cobra.Command) *cobra.Command {
	var sshUser string

	cmd := &cobra.Command{
		Use:   "scp [VM_ID:]SOURCE [VM_ID:]DESTINATION",
		Short: "Copy files to/from a VM",
		Long: `Securely copy files to or from a VM using SCP.

This command allows you to copy files between your local machine and a VM. You can specify the VM ID followed by a colon and the path on the VM, or just a local path.

To copy a file from your local machine to a VM, use:
replicated vm scp localfile.txt vm-id:/path/on/vm/

To copy a file from a VM to your local machine, use:
replicated vm scp vm-id:/path/on/vm/file.txt localfile.txt

If no VM ID is provided and multiple VMs are available, you will be prompted to select a VM.

The SSH user can be specified in order of precedence:
1. By specifying the -u flag
2. REPLICATED_SSH_USER environment variable
3. GITHUB_ACTOR environment variable (from GitHub Actions)
4. GITHUB_USER environment variable

Note: Only running VMs can be connected to via SCP.`,
		Example: `# Copy a local file to a VM
replicated vm scp localfile.txt vm-id:/home/user/

# Copy a file from a VM to local machine
replicated vm scp vm-id:/home/user/file.txt localfile.txt

# Copy with a specific user
replicated vm scp -u myuser localfile.txt vm-id:/home/myuser/

# Interactive VM selection (if VM ID is not specified)
replicated vm scp localfile.txt :/home/user/`,
		RunE: r.scpVM,
		Args: cobra.ExactArgs(2),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&sshUser, "user", "u", "", "SSH user to connect with")

	return cmd
}

func (r *runners) scpVM(cmd *cobra.Command, args []string) error {
	if err := r.initVMClient(); err != nil {
		return err
	}

	sshUser, _ := cmd.Flags().GetString("user")
	source := args[0]
	destination := args[1]

	// Parse source and destination to determine if they contain VM IDs
	sourceVMID, sourcePath, sourceIsRemote := parseScpPath(source)
	destVMID, destPath, destIsRemote := parseScpPath(destination)

	// Ensure we're not trying to copy between two VMs
	if sourceIsRemote && destIsRemote {
		return errors.New("copying between two VMs is not supported. Please copy to your local machine first")
	}

	// Ensure we're copying either to or from a VM
	if !sourceIsRemote && !destIsRemote {
		return errors.New("at least one of source or destination must be a VM path (prefixed with VM_ID:)")
	}

	// Handle VM selection and validation for source
	if sourceIsRemote {
		var err error
		sourceVMID, err = validateOrSelectVM(r, sourceVMID, "Select source VM for SCP transfer")
		if err != nil {
			return err
		}
	}

	// Handle VM selection and validation for destination
	if destIsRemote {
		var err error
		destVMID, err = validateOrSelectVM(r, destVMID, "Select destination VM for SCP transfer")
		if err != nil {
			return err
		}
	}

	// Execute the appropriate SCP command
	if sourceIsRemote {
		// Copy from VM to local
		return r.kotsAPI.SCPFromVM(sourceVMID, sshUser, sourcePath, destPath)
	} else {
		// Copy from local to VM
		return r.kotsAPI.SCPToVM(destVMID, sshUser, sourcePath, destPath)
	}
}

// validateOrSelectVM validates a VM ID if provided, or prompts the user to select a VM
// Returns the validated or selected VM ID
func validateOrSelectVM(r *runners, vmID string, selectionPrompt string) (string, error) {
	if vmID != "" {
		// Check VM status if ID is provided
		vm, err := r.kotsAPI.GetVM(vmID)
		if err != nil {
			return "", errors.Wrap(err, "failed to get VM")
		}
		if !isVMRunning(vm) {
			return "", fmt.Errorf("VM %s is not running (current status: %s). Cannot connect to a non-running VM", vmID, vm.Status)
		}
		return vmID, nil
	}

	// Get all VMs
	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to list VMs")
	}

	// Only show running VMs
	runningVMs := filterVMsByStatus(vms, types.VMStatusRunning)
	if len(runningVMs) == 0 {
		return "", handleNoRunningVMs(vms)
	}

	// Always prompt for VM selection, even if there's only one VM
	selectedVM, err := selectVM(runningVMs, selectionPrompt)
	if err != nil {
		return "", errors.Wrap(err, "failed to select VM")
	}
	return selectedVM.ID, nil
}

// parseScpPath parses a path string to determine if it contains a VM ID
// Returns vmID, path, and whether the path is remote (on a VM)
func parseScpPath(path string) (string, string, bool) {
	for i, c := range path {
		if c == ':' {
			if i == 0 {
				// Path starts with a colon, indicating an empty VM ID (for selection)
				return "", path[i+1:], true
			}
			return path[:i], path[i+1:], true
		}
	}
	return "", path, false
}
