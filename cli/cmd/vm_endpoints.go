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
	if endpointType == "scp" {
		outputFormat = "format for use with scp command"
	}

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

	return r.getVMEndpoint(args[0], endpointType)
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

	// Map of endpoint formatters based on endpoint type
	formatters := map[string]func(string, string, int64) string{
		"ssh": func(username, host string, port int64) string {
			return fmt.Sprintf("ssh://%s@%s:%d", username, host, port)
		},
		"scp": func(username, host string, port int64) string {
			return fmt.Sprintf("-P %d %s@%s:", port, username, host)
		},
	}

	// Get the formatter for the requested endpoint type
	formatter, ok := formatters[endpointType]
	if !ok {
		return errors.Errorf("unsupported endpoint type: %s", endpointType)
	}

	// Format and print the endpoint
	fmt.Println(formatter(githubUsername, vm.DirectSSHEndpoint, vm.DirectSSHPort))

	return nil
}
