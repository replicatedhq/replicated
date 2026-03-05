package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshot(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage VM snapshots.",
		Long: `The 'vm snapshot' command is a parent command for managing snapshots of CMX VMs. It allows users to create, list, remove, and restore snapshots. Use the subcommands to manage VM snapshot operations.

Snapshots capture the full state of a running VM, allowing you to restore to a known-good state later. The restore command creates a new VM from a snapshot using the original VM's configuration.`,
		Example: `# List snapshots for a VM
replicated vm snapshot ls --vm-id VM_ID

# Create a snapshot of a running VM
replicated vm snapshot create --vm-id VM_ID

# Remove a snapshot
replicated vm snapshot rm --vm-id VM_ID SNAPSHOT_ID

# Restore a snapshot (creates a new VM)
replicated vm snapshot restore --vm-id VM_ID SNAPSHOT_ID`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}
