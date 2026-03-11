package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotRestore(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore SNAPSHOT_ID",
		Short: "Restore a VM from a snapshot.",
		Long: `The 'vm snapshot restore' command creates a new VM from a snapshot. The new VM is created using the original VM's configuration (distribution, version, instance type, disk size, etc.).

The snapshot must be in a 'ready' state for restore to succeed. The command returns the newly created VM.`,
		Example: `# Restore a VM from a snapshot
replicated vm snapshot restore --vm-id VM_ID SNAPSHOT_ID

# Restore with a custom TTL
replicated vm snapshot restore --vm-id VM_ID SNAPSHOT_ID --ttl 2h

# Restore and output in JSON format
replicated vm snapshot restore --vm-id VM_ID SNAPSHOT_ID --output json`,
		RunE: r.vmSnapshotRestore,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotVMID, "vm-id", "", "The ID of the VM to restore (required)")
	err := cmd.MarkFlagRequired("vm-id")
	if err != nil {
		panic(err)
	}
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeTerminatedVMIDs)
	cmd.ValidArgsFunction = r.completeVMSnapshotIDs
	cmd.Flags().StringVar(&r.args.vmSnapshotTTL, "ttl", "", "VM TTL (duration)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotRestore(_ *cobra.Command, args []string) error {
	snapshotID := args[0]

	vm, err := r.kotsAPI.RestoreVMSnapshot(r.args.vmSnapshotVMID, snapshotID, r.args.vmSnapshotTTL)
	if err != nil {
		return errors.Wrap(err, "restore vm snapshot")
	}

	return print.VM(r.outputFormat, r.w, vm)
}
