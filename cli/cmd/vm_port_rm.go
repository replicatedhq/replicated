package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMPortRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm VM_ID_OR_NAME --id PORT_ID",
		Short: "Remove vm port by ID.",
		Long: `The 'vm port rm' command removes a specific port from a vm. You must provide the ID or name of the vm and the ID of the port to remove.

This command is useful for managing the network settings of your test vms by allowing you to clean up unused or incorrect ports. After removing a port, the updated list of ports will be displayed.`,
		Example: `# Remove a port using its ID
replicated vm port rm VM_ID_OR_NAME --id PORT_ID

# Remove a port and display the result in JSON format
replicated vm port rm VM_ID_OR_NAME --id PORT_ID --output json`,
		RunE:              r.vmPortRemove,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmPortRemoveAddonID, "id", "", "ID of the port to remove (required)")
	err := cmd.MarkFlagRequired("id")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmPortRemove(_ *cobra.Command, args []string) error {
	vmID, err := r.getVMIDFromArg(args[0])
	if err != nil {
		return err
	}

	err = r.kotsAPI.DeleteVMAddon(vmID, r.args.vmPortRemoveAddonID)
	if err != nil {
		return err
	}

	ports, err := r.kotsAPI.ListVMPorts(vmID)
	if err != nil {
		return err
	}

	return print.VMPorts(r.outputFormat, ports, true)
}
