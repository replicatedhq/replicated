package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSSHEndpoint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-endpoint VM_ID",
		Short: "Get the SSH endpoint of a VM",
		Long: `Get the SSH endpoint and port of a VM.

The output will be in the format: hostname:port`,
		Example: `# Get SSH endpoint for a specific VM by ID
replicated vm ssh-endpoint <id>`,
		RunE:              r.VMSSHEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) VMSSHEndpoint(cmd *cobra.Command, args []string) error {
	return r.getVMEndpoint(args[0], "ssh")
}

// getVMEndpoint retrieves and formats VM endpoint with the specified protocol
// endpointType should be either "ssh" or "scp"
func (r *runners) getVMEndpoint(vmID, endpointType string) error {
	vm, err := r.kotsAPI.GetVM(vmID)
	if err != nil {
		return errors.Wrap(err, "get vm")
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", vm.ID)
	}

	// Get GitHub username from API
	githubUsername, err := r.kotsAPI.GetGitHubUsername()
	if err != nil {
		return errors.Wrap(err, "get github username")
	}

	// Format the endpoint with username if available
	if githubUsername == "" {
		return errors.New(`no github account associated with vendor portal user
Visit https://vendor.replicated.com/account-settings to link your account`)
	}

	if endpointType == "ssh" {
		fmt.Printf("ssh://%s@%s:%d\n", githubUsername, vm.DirectSSHEndpoint, vm.DirectSSHPort)
	} else if endpointType == "scp" {
		fmt.Printf("-P %d %s@%s:\n", vm.DirectSSHPort, githubUsername, vm.DirectSSHEndpoint)
	}

	return nil
}
