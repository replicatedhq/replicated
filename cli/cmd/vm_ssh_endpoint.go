package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSSHEndpoint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-endpoint VM_ID_OR_NAME",
		Short: "Get the SSH endpoint of a VM",
		Long: `Get the SSH endpoint and port of a VM.

The output will be in the format: hostname:port

You can identify the VM either by its unique ID or by its name.

Note: SSH endpoints can only be retrieved from VMs in the "running" state.`,
		Example: `# Get SSH endpoint for a specific VM by ID
replicated vm ssh-endpoint aaaaa11

# Get SSH endpoint for a specific VM by name
replicated vm ssh-endpoint my-test-vm`,
		RunE:              r.VMSSHEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) VMSSHEndpoint(cmd *cobra.Command, args []string) error {
	vmID, err := r.getVMIDFromArg(args[0])
	if err != nil {
		return err
	}

	vm, err := r.kotsAPI.GetVM(vmID)
	if err != nil {
		return errors.Wrap(err, "get vm")
	}

	if vm.Status != "running" {
		return errors.Errorf("VM %s is not in running state (current state: %s). SSH is only available for running VMs", vm.ID, vm.Status)
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", vm.ID)
	}

	// Get GitHub username from API
	githubUsername, err := r.kotsAPI.GetGitHubUsername()
	if err != nil {
		return errors.Wrap(err, "get github username")
	}

	// Format the SSH endpoint with username if available
	if githubUsername == "" {
		return errors.New(`no github account associated with vendor portal user
Visit https://vendor.replicated.com/account-settings to link your account`)
	}

	fmt.Printf("ssh://%s@%s:%d\n", githubUsername, vm.DirectSSHEndpoint, vm.DirectSSHPort)

	return nil
}

func (r *runners) getVMIDFromArg(arg string) (string, error) {
	_, err := r.kotsAPI.GetVM(arg)
	if err == nil {
		return arg, nil
	}

	cause := errors.Cause(err)
	if cause != platformclient.ErrNotFound && cause != platformclient.ErrForbidden {
		return "", errors.Wrap(err, "get vm")
	}

	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return "", ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return "", errors.Wrap(err, "list vms")
	}

	var matchingVMs []string
	for _, vm := range vms {
		if vm.Name == arg {
			matchingVMs = append(matchingVMs, vm.ID)
		}
	}

	switch len(matchingVMs) {
	case 0:
		return "", errors.Errorf("VM with name or ID %q not found", arg)
	case 1:
		return matchingVMs[0], nil
	default:
		return "", errors.Errorf("Multiple VMs found with name %q. Please use the VM ID instead. Matching VMs: %s. To view all VM IDs run `replicated vm ls`",
			arg,
			fmt.Sprintf("%s (and %d more)", matchingVMs[0], len(matchingVMs)-1))
	}
}
