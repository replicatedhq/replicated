package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// InitVMSSHEndpoint initializes the command to get SSH endpoint
func (r *runners) InitVMSSHEndpoint(parent *cobra.Command) *cobra.Command {
	return r.initVMEndpointCmd(parent, "ssh")
}

// InitVMSCPEndpoint initializes the command to get SCP endpoint
func (r *runners) InitVMSCPEndpoint(parent *cobra.Command) *cobra.Command {
	return r.initVMEndpointCmd(parent, "scp")
}

// initVMEndpointCmd creates a command for either SSH or SCP endpoint retrieval
func (r *runners) initVMEndpointCmd(parent *cobra.Command, endpointType string) *cobra.Command {
	protocol := strings.ToUpper(endpointType)
	cmdUse := fmt.Sprintf("%s-endpoint VM_ID", endpointType)
	cmdShort := fmt.Sprintf("Get the %s endpoint of a VM", protocol)

	outputFormat := "hostname:port"

	cmdLong := fmt.Sprintf(`Get the %s endpoint and port of a VM.

The output will be in the %s`, protocol, outputFormat)

	cmdExample := fmt.Sprintf(`# Get %s endpoint for a specific VM by ID
replicated vm %s-endpoint <id>`, protocol, endpointType)

	cmd := &cobra.Command{
		Use:               cmdUse,
		Short:             cmdShort,
		Long:              cmdLong,
		Example:           cmdExample,
		RunE:              r.VMEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

// VMEndpoint handles both the ssh-endpoint and scp-endpoint commands
func (r *runners) VMEndpoint(cmd *cobra.Command, args []string) error {
	// Determine endpoint type based on command name
	endpointType := "ssh"
	if strings.Contains(cmd.Use, "scp-endpoint") {
		endpointType = "scp"
	}

	return r.getVMEndpoint(args[0], endpointType, nil, "")
}

// getVMEndpoint retrieves and formats VM endpoint with the specified protocol
// endpointType should be either "ssh" or "scp"
type VM struct {
	DirectSSHEndpoint string
	DirectSSHPort     int64
	ID                string
}

func (r *runners) getVMEndpoint(vmID, endpointType string, vm *VM, githubUsername string) error {
	var err error

	// Use vm if provided, otherwise fetch from API
	if vm == nil {
		vmFromAPI, err := r.kotsAPI.GetVM(vmID)
		if err != nil {
			return errors.Wrap(err, "get vm")
		}
		vm = &VM{
			DirectSSHEndpoint: vmFromAPI.DirectSSHEndpoint,
			DirectSSHPort:     vmFromAPI.DirectSSHPort,
			ID:                vmFromAPI.ID,
		}
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
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
		return errors.New(`no github account associated with vendor portal user
Visit https://vendor.replicated.com/account-settings to link your account`)
	}

	// Format the endpoint URL with the appropriate protocol
	fmt.Printf("%s://%s@%s:%d\n", endpointType, githubUsername, vm.DirectSSHEndpoint, vm.DirectSSHPort)

	return nil
}
