package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMGetSSHEndpoint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-endpoint VM_ID",
		Short: "Get the SSH endpoint of a VM",
		Long: `Get the SSH endpoint and port of a VM.

The output will be in the format: hostname:port`,
		Example: `# Get SSH endpoint for a specific VM by ID
replicated vm get ssh-endpoint <id>`,
		RunE:              r.getVMSSHEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) getVMSSHEndpoint(cmd *cobra.Command, args []string) error {
	vmID := args[0]

	vms, err := r.kotsAPI.ListVMs(false, nil, nil)
	if err != nil {
		return errors.Wrap(err, "list vms")
	}

	var vm *types.VM
	for _, v := range vms {
		if v.ID == vmID {
			vm = v
			break
		}
	}
	if vm == nil {
		return errors.Errorf("VM %s not found", vmID)
	}

	if vm.DirectSSHEndpoint == "" || vm.DirectSSHPort == 0 {
		return errors.Errorf("VM %s does not have SSH endpoint configured", vm.ID)
	}

	fmt.Printf("ssh://%s:%d\n", vm.DirectSSHEndpoint, vm.DirectSSHPort)
	return nil
}
