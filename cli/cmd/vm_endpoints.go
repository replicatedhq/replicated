package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

const githubAccountSettingsURL = "https://vendor.replicated.com/account-settings"

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

	cmdLong := fmt.Sprintf(`Get the %s endpoint and port of a VM.

The output will be in the format: %s

You can identify the VM either by its unique ID or by its name.

Note: %s endpoints can only be retrieved from VMs in the "running" state.`, protocol, outputFormat, protocol)

	cmdExample := fmt.Sprintf(`# Get %s endpoint for a specific VM by ID
replicated vm %s-endpoint aaaaa11

# Get %s endpoint for a specific VM by name
replicated vm %s-endpoint my-test-vm`, protocol, endpointType, protocol, endpointType)

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

	// Validate endpoint type
	if err := validateEndpointType(endpointType); err != nil {
		return err
	}

	vmID, err := r.getVMIDFromArg(args[0])
	if err != nil {
		return err
	}

	return r.getVMEndpoint(vmID, endpointType, nil, "")
}

// getVMEndpoint retrieves and formats VM endpoint with the specified protocol
// endpointType should be either "ssh" or "scp"
func (r *runners) getVMEndpoint(vmID, endpointType string, vm *VM, githubUsername string) error {
	var err error

	// Use vm if provided, otherwise fetch from API
	if vm == nil {
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
	}

	if vm.Status != "running" {
		return errors.Errorf("VM %s is not in running state (current state: %s). %s is only available for running VMs",
			vm.ID, vm.Status, strings.ToUpper(endpointType))
	}

	if vm.DirectEndpoint == "" || vm.DirectPort == 0 {
		return errors.Errorf("VM %s does not have %s endpoint configured", vm.ID, endpointType)
	}

	// if kotsAPI is not nil and githubUsername is not provided, fetch from API
	if r.kotsAPI != nil && githubUsername == "" {
		// Get GitHub username from API
		githubUsername, err = r.kotsAPI.GetGitHubUsername()
		if err != nil {
			return errors.Wrap(err, "get github username")
		}
	}

	// Format the endpoint with username if available
	if githubUsername == "" {
		return errors.Errorf(`no github account associated with vendor portal user
Visit %s to link your account`, githubAccountSettingsURL)
	}

	// Format the endpoint URL with the appropriate protocol
	fmt.Printf("%s://%s@%s:%d\n", endpointType, githubUsername, vm.DirectEndpoint, vm.DirectPort)

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
