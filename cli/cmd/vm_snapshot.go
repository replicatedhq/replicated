package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshot(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage VM snapshots.",
		Long: `The 'vm snapshot' command is a parent command for managing snapshots of CMX VMs. It allows users to create, list, remove, update, and restore snapshots. Use the subcommands to manage VM snapshot operations.

Snapshots capture the full state of a running VM, allowing you to restore to a known-good state later. The restore command creates a new VM from a snapshot using the original VM's configuration.

Commands that take a snapshot accept either SNAPSHOT_ID (or short id) as a positional argument or --name to specify by snapshot name.

VM snapshots are currently an alpha feature.`,
		Hidden: true,
		Example: `# List all snapshots
replicated vm snapshot ls

# Create a snapshot of a running VM
replicated vm snapshot create --vm-id VM_ID

# Remove a snapshot by ID or name
replicated vm snapshot rm SNAPSHOT_ID
replicated vm snapshot rm --name "my-snapshot"

# Update snapshot TTL
replicated vm snapshot update SNAPSHOT_ID --ttl 48h
replicated vm snapshot update --name "my-snapshot" --ttl 48h

# Restore a snapshot (creates a new VM)
replicated vm snapshot restore SNAPSHOT_ID
replicated vm snapshot restore --name "my-snapshot"`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.PersistentFlags().StringVar(&r.args.vmSnapshotName, "name", "", "Snapshot name (alternative to SNAPSHOT_ID)")
	cmd.RegisterFlagCompletionFunc("name", r.completeVMSnapshotNames)

	return cmd
}
