package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSCPEndpoint(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scp-endpoint VM_ID",
		Short: "Get the SCP endpoint of a VM",
		Long: `Get the SCP endpoint and port of a VM.

The output will be in the format: hostname:port`,
		Example: `# Get SCP endpoint for a specific VM by ID
replicated vm scp-endpoint <id>`,
		RunE:              r.VMSCPEndpoint,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) VMSCPEndpoint(cmd *cobra.Command, args []string) error {
	return r.getVMEndpoint(args[0], "scp")
}
