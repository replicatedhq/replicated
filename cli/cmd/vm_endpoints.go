package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

const (
	EndpointTypeSSH = "ssh"
	EndpointTypeSCP = "scp"
)

type VM struct {
	DirectEndpoint string
	DirectPort     int64
	ID             string
	Status         string
	Name           string
}

// validateEndpointType validates that the endpoint type is supported
func validateEndpointType(endpointType string) error {
	if endpointType != EndpointTypeSSH && endpointType != EndpointTypeSCP {
		return errors.Errorf("invalid endpoint type: %s", endpointType)
	}
	return nil
}

// InitVMSSHEndpoint initializes the command to get SSH endpoint
func (r *runners) InitVMSSHEndpoint(parent *cobra.Command) *cobra.Command {
	return r.initVMEndpointCmd(parent, EndpointTypeSSH)
}

// InitVMSCPEndpoint initializes the command to get SCP endpoint
func (r *runners) InitVMSCPEndpoint(parent *cobra.Command) *cobra.Command {
	return r.initVMEndpointCmd(parent, EndpointTypeSCP)
}

// initVMEndpointCmd creates a command for either SSH or SCP endpoint retrieval
func (r *runners) initVMEndpointCmd(parent *cobra.Command, endpointType string) *cobra.Command {
	protocol := strings.ToUpper(endpointType)
	cmdUse := fmt.Sprintf("%s-endpoint VM_ID_OR_NAME", endpointType)
	cmdShort := fmt.Sprintf("Get the %s endpoint of a VM", protocol)
	outputFormat := fmt.Sprintf("%s://username@hostname:port", endpointType)

	var scriptExample string
	switch endpointType {
	case EndpointTypeSSH:
		scriptExample = `# Use the endpoint to SSH to a VM by name
ssh $(replicated vm ssh-endpoint my-test-vm)`
	case EndpointTypeSCP:
		scriptExample = `# Use the endpoint to SCP a file to a VM by name
scp /tmp/my-file $(replicated vm scp-endpoint my-test-vm)//dst/path/my-file

# Use the endpoint to SCP a file from a VM by name
scp $(replicated vm scp-endpoint my-test-vm)//src/path/my-file /tmp/my-file`
	default:
		scriptExample = ""
	}

	cmdLong := fmt.Sprintf(`Get the %s endpoint and port of a VM.

The output will be in the format: %s

You can identify the VM either by its unique ID or by its name.

Note: %s endpoints can only be retrieved from VMs in the "running" state.

VMs are currently a beta feature.`, protocol, outputFormat, protocol)

	cmdExample := fmt.Sprintf(`# Get %s endpoint for a specific VM by ID
replicated vm %s-endpoint aaaaa11

# Get %s endpoint for a specific VM by name
replicated vm %s-endpoint my-test-vm

# Get %s endpoint with a custom username
replicated vm %s-endpoint my-test-vm --username custom-user

%s`, protocol, endpointType, protocol, endpointType, protocol, endpointType, scriptExample)

	cmd := &cobra.Command{
		Use:               cmdUse,
		Short:             cmdShort,
		Long:              cmdLong,
		Example:           cmdExample,
		RunE:              r.VMEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDsAndNames,
		Annotations:       map[string]string{"endpointType": endpointType},
	}

	cmd.Flags().String("username", "", fmt.Sprintf("Custom username to use in %s endpoint instead of the GitHub username set in Vendor Portal", protocol))

	parent.AddCommand(cmd)

	return cmd
}

// VMEndpoint handles both the ssh-endpoint and scp-endpoint commands
func (r *runners) VMEndpoint(cmd *cobra.Command, args []string) error {
	// Get endpoint type from command annotation
	endpointType := EndpointTypeSSH // Default fallback
	if cmd.Annotations != nil {
		if epType, ok := cmd.Annotations["endpointType"]; ok {
			endpointType = epType
		}
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return errors.Wrap(err, "get username")
	}

	vmID, err := r.getVMIDFromArg(args[0])
	if err != nil {
		return err
	}

	return r.getVMEndpoint(vmID, endpointType, username)
}

// getVMEndpoint retrieves and formats VM endpoint with the specified protocol
// endpointType should be either "ssh" or "scp"
func (r *runners) getVMEndpoint(vmID, endpointType, username string) error {
	var err error
	var vm *VM

	// Validate endpoint type
	if err := validateEndpointType(endpointType); err != nil {
		return err
	}

	vmFromAPI, err := r.kotsAPI.GetVM(vmID)
	if err != nil {
		return errors.Wrap(err, "get vm")
	}
	vm = &VM{
		DirectEndpoint: vmFromAPI.DirectSSHEndpoint,
		DirectPort:     vmFromAPI.DirectSSHPort,
		ID:             vmFromAPI.ID,
		Status:         string(vmFromAPI.Status),
	}

	if vm.Status != "running" {
		return errors.Errorf("VM %s is not in running state (current state: %s). %s is only available for running VMs",
			vm.ID, vm.Status, strings.ToUpper(endpointType))
	}

	if vm.DirectEndpoint == "" || vm.DirectPort == 0 {
		return errors.Errorf("VM %s does not have %s endpoint configured", vm.ID, endpointType)
	}

	// Use provided username or fetch from GitHub
	if username == "" {
		githubUsername, err := r.kotsAPI.GetGitHubUsername()
		if err != nil {
			return errors.Wrap(errors.New(`failed to obtain GitHub username from Vendor Portal
Alternatively, you can use the --username flag to specify a custom username for the endpoint`), "get github username")
		}

		if githubUsername == "" {
			return errors.Errorf(`no GitHub account associated with Vendor Portal user
Visit the Account Settings page in Vendor Portal to link your account
Alternatively, you can use the --username flag to specify a custom username for the endpoint`)
		}

		username = githubUsername
	}

	// Format the endpoint URL with the appropriate protocol
	fmt.Printf("%s://%s@%s:%d\n", endpointType, username, vm.DirectEndpoint, vm.DirectPort)

	return nil
}

// getVMIDFromArg resolves a VM ID or name to a VM ID
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
		return "", errors.Errorf("VM with name or ID '%q' not found", arg)
	case 1:
		return matchingVMs[0], nil
	default:
		return "", errors.Errorf("Multiple VMs found with name '%q'. Please use the VM ID instead. Matching VMs: %s. To view all VM IDs run `replicated vm ls`",
			arg,
			fmt.Sprintf("%s (and %d more)", matchingVMs[0], len(matchingVMs)-1))
	}
}
