package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotRestore(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore [SNAPSHOT_ID]",
		Short: "Restore a VM from a snapshot.",
		Long: `The 'vm snapshot restore' command creates a new VM from a snapshot. The new VM is created using the original VM's configuration (distribution, version, instance type, disk size, etc.).

The snapshot must be in a 'ready' state for restore to succeed. The command returns the newly created VM. Provide SNAPSHOT_ID (or short id) or use --name to specify by snapshot name.

VM snapshots are currently an alpha feature.`,
		Example: `# Restore a VM from a snapshot
replicated vm snapshot restore SNAPSHOT_ID

# Restore by name with a custom TTL
replicated vm snapshot restore --name "my-snapshot" --ttl 2h

# Restore and output in JSON format
replicated vm snapshot restore SNAPSHOT_ID --output json`,
		RunE: r.vmSnapshotRestore,
		Args: cobra.MaximumNArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.ValidArgsFunction = r.completeVMSnapshotIDsAndNames
	cmd.Flags().StringVar(&r.args.vmSnapshotTTL, "ttl", "", "VM TTL (duration)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotRestore(_ *cobra.Command, args []string) error {
	snapshotIDOrName, err := r.ensureSnapshotIDArg(args)
	if err != nil {
		return err
	}

	vm, err := r.kotsAPI.RestoreVMSnapshot(snapshotIDOrName, r.args.vmSnapshotTTL)
	if err != nil {
		return errors.Wrap(err, "restore vm snapshot")
	}

	return print.VM(r.outputFormat, r.w, vm)
}
