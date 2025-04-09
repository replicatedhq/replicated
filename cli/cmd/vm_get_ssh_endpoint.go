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
		RunE:              r.getVMSSHEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) getVMSSHEndpoint(cmd *cobra.Command, args []string) error {
	vmID := args[0]

	vm, err := r.kotsAPI.GetVM(vmID)
	if err != nil {
		return errors.Wrap(err, "get vm")
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", vm.ID)
	}

	fmt.Printf("ssh://%s:%d\n", vm.DirectSSHEndpoint, vm.DirectSSHPort)
	return nil
}
