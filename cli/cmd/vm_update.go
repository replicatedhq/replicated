package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMUpdateCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update VM settings.",
		Long: `The 'vm update' command allows you to modify the settings of a virtual machine. You can update a VM either by providing its ID or name directly as an argument to subcommands, or by using the '--id' or '--name' flags. This command supports updating various VM settings, which will be handled by specific subcommands.

- To update the VM by its ID or name, use the subcommand directly with the ID or name as an argument.
- Alternatively, to update the VM by its ID, use the '--id' flag.
- Alternatively, to update the VM by its name, use the '--name' flag.

Subcommands will allow for more specific updates like TTL.

VMs are currently a beta feature.`,
		Example: `# Update a VM TTL by specifying its ID or name directly
replicated vm update ttl my-test-vm --ttl 12h

# Update a VM by specifying its ID with a flag
replicated vm update --id aaaaa11 --ttl 12h

# Update a VM by specifying its name with a flag
replicated vm update --name my-test-vm --ttl 12h`,
	}
	parent.AddCommand(cmd)

	cmd.PersistentFlags().StringVar(&r.args.updateVMName, "name", "", "Name of the vm to update.")
	cmd.RegisterFlagCompletionFunc("name", r.completeVMNames)

	cmd.PersistentFlags().StringVar(&r.args.updateVMID, "id", "", "id of the vm to update (when name is not provided)")
	cmd.RegisterFlagCompletionFunc("id", r.completeVMIDs)

	return cmd
}

func (r *runners) ensureUpdateVMIDArg(args []string) error {
	// by default, we look at args[0] as the id
	// but if it's not provided, we look for a viper flag named "name" and use it
	// as the name of the vm, not the id
	if len(args) > 0 {
		r.args.updateVMID = args[0]
	} else if r.args.updateVMName != "" {
		vms, err := r.kotsAPI.ListVMs(false, nil, nil)
		if errors.Cause(err) == platformclient.ErrForbidden {
			return ErrCompatibilityMatrixTermsNotAccepted
		} else if err != nil {
			return errors.Wrap(err, "list vms")
		}
		for _, vm := range vms {
			if vm.Name == r.args.updateVMName {
				r.args.updateVMID = vm.ID
				break
			}
		}
	} else if r.args.updateVMID != "" {
		// do nothing
		// but this is here for readability
	} else {
		return errors.New("must provide vm id or name")
	}

	return nil
}
