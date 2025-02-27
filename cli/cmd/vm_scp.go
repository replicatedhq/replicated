package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
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

	// Handle the case where we need to select a VM
	if sourceIsRemote && sourceVMID == "" {
		selectedVM, err := selectRunningVM(r.kotsAPI)
		if err != nil {
			return err
		}
		sourceVMID = selectedVM.ID
	}

	if destIsRemote && destVMID == "" {
		selectedVM, err := selectRunningVM(r.kotsAPI)
		if err != nil {
			return err
		}
		destVMID = selectedVM.ID
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

// selectRunningVM prompts the user to select a running VM
func selectRunningVM(kotsAPI *kotsclient.VendorV3Client) (*types.VM, error) {
	vms, err := kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list VMs")
	}

	// Filter to only show running VMs
	runningVMs := filterVMsByStatus(vms, types.VMStatusRunning)
	if len(runningVMs) == 0 {
		// Show information about non-running VMs if they exist
		nonTerminatedVMs := filterVMsByStatus(vms, "")
		if len(nonTerminatedVMs) == 0 {
			return nil, errors.New("no active VMs found")
		}

		// List non-running VMs with their statuses
		fmt.Println("No running VMs found. The following VMs are available but not running:")
		for _, vm := range nonTerminatedVMs {
			if vm.Status != types.VMStatusTerminated {
				fmt.Printf("  - %s (ID: %s, Status: %s)\n", vm.Name, vm.ID, vm.Status)
			}
		}
		return nil, errors.New("SCP connection requires a running VM. Please start a VM before connecting")
	}

	// Always prompt for VM selection, even if there's only one VM
	return selectVM(runningVMs)
}
