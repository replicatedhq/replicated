package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMPortLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls VM_ID_OR_NAME",
		Short: "List vm ports for a vm.",
		Long: `The 'vm port ls' command lists all the ports configured for a specific vm. You must provide the vm ID or name to retrieve and display the ports.

This command is useful for viewing the current port configurations, protocols, and other related settings of your test vm. The output format can be customized to suit your needs, and the available formats include table, JSON, and wide views.`,
		Example: `# List ports for a vm in the default table format
replicated vm port ls VM_ID_OR_NAME

# List ports for a vm in JSON format
replicated vm port ls VM_ID_OR_NAME --output json

# List ports for a vm in wide format
replicated vm port ls VM_ID_OR_NAME --output wide`,
		RunE:              r.vmPortList,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmPortList(_ *cobra.Command, args []string) error {
	vmID, err := r.getVMIDFromArg(args[0])
	if err != nil {
		return err
	}

	ports, err := r.kotsAPI.ListVMPorts(vmID)
	if err != nil {
		return err
	}

	return print.VMPorts(r.outputFormat, ports, true)
}
