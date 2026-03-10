package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of a running VM.",
		Long: `The 'vm snapshot create' command creates a snapshot of a running VM. The VM must be in a running state for snapshot creation to succeed.

IMPORTANT: The VM will be temporarily paused during snapshot creation. This may result in a brief service interruption or pause in VM activity.

The snapshot is created asynchronously. The command returns immediately with the snapshot in a pending state. Use 'vm snapshot ls' to check the snapshot status.`,
		Example: `# Create a snapshot of a VM
replicated vm snapshot create --vm-id VM_ID

# Create a named snapshot
replicated vm snapshot create --vm-id VM_ID --name "before-upgrade"

# Create a snapshot and output in JSON format
replicated vm snapshot create --vm-id VM_ID --output json`,
		RunE: r.vmSnapshotCreate,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotVMID, "vm-id", "", "The ID of the VM to snapshot (required)")
	err := cmd.MarkFlagRequired("vm-id")
	if err != nil {
		panic(err)
	}
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeVMIDs)
	cmd.Flags().StringVar(&r.args.vmSnapshotCreateName, "name", "", "Optional name for the snapshot")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotCreate(_ *cobra.Command, args []string) error {
	snapshot, err := r.kotsAPI.CreateVMSnapshot(r.args.vmSnapshotVMID, r.args.vmSnapshotCreateName)
	if err != nil {
		return errors.Wrap(err, "create vm snapshot")
	}

	return print.VMSnapshot(r.outputFormat, r.w, snapshot)
}
