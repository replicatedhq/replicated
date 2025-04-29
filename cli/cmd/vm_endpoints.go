package cmd

import (
	"fmt"
	"reflect"
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
func (r *runners) getVMEndpoint(vmID, endpointType string, vm interface{}, githubUsername string) error {
	var err error
	var directSSHEndpoint string
	var directSSHPort int
	var id string

	// Use vm if provided, otherwise fetch from API
	if vm != nil {
		// Extract VM fields from vm (map type)
		if vmMap, ok := vm.(map[string]interface{}); ok {
			directSSHEndpoint, _ = vmMap["DirectSSHEndpoint"].(string)
			directSSHPort, _ = vmMap["DirectSSHPort"].(int)
			id, _ = vmMap["ID"].(string)
		} else {
			return errors.New("unexpected VM type")
		}
	} else {
		vm, err = r.kotsAPI.GetVM(vmID)
		if err != nil {
			return errors.Wrap(err, "get vm")
		}

		// Extract VM fields based on type
		switch typedVM := vm.(type) {
		case map[string]interface{}:
			directSSHEndpoint, _ = typedVM["DirectSSHEndpoint"].(string)
			directSSHPort, _ = typedVM["DirectSSHPort"].(int)
			id, _ = typedVM["ID"].(string)
		default:
			// Use reflection to access fields for any struct type
			vmValue := reflect.ValueOf(vm)
			if vmValue.Kind() == reflect.Ptr {
				vmValue = vmValue.Elem()
			}

			if vmValue.Kind() == reflect.Struct {
				// Try to extract fields by name
				idField := vmValue.FieldByName("ID")
				if idField.IsValid() && idField.Kind() == reflect.String {
					id = idField.String()
				}

				endpointField := vmValue.FieldByName("DirectSSHEndpoint")
				if endpointField.IsValid() && endpointField.Kind() == reflect.String {
					directSSHEndpoint = endpointField.String()
				}

				portField := vmValue.FieldByName("DirectSSHPort")
				if portField.IsValid() && portField.Kind() == reflect.Int {
					directSSHPort = int(portField.Int())
				}
			} else {
				return errors.New("unable to extract VM fields from response")
			}
		}
	}

	if directSSHEndpoint == "" || directSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", id)
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
	fmt.Printf("%s://%s@%s:%d\n", endpointType, githubUsername, directSSHEndpoint, directSSHPort)

	return nil
}
