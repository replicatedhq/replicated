package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMPort(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port",
		Short: "Manage VM ports.",
		Long: `The 'vm port' command is a parent command for managing ports in a vm. It allows users to list, remove, or expose specific ports used by the vm. Use the subcommands (such as 'ls', 'rm', and 'expose') to manage port configurations effectively.

This command provides flexibility for handling ports in various test vms, ensuring efficient management of vm networking settings.`,
		Example: `# List all exposed ports in a vm
replicated vm port ls VM_ID_OR_NAME

# Remove an exposed port from a vm
replicated vm port rm VM_ID_OR_NAME --id PORT_ID

# Expose a new port in a vm
replicated vm port expose VM_ID_OR_NAME --port PORT`,
		SilenceUsage: true,
		Hidden:       false,
	}
	parent.AddCommand(cmd)

	return cmd
}
