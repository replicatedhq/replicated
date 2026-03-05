package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a snapshot of a running VM.",
		Long: `The 'vm snapshot create' command creates a snapshot of a running VM. The VM must be in a running state for snapshot creation to succeed.

The snapshot is created asynchronously. The command returns immediately with the snapshot in a pending state. Use 'vm snapshot ls' to check the snapshot status.`,
		Example: `# Create a snapshot of a VM
replicated vm snapshot create --vm-id VM_ID

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
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotCreate(_ *cobra.Command, args []string) error {
	snapshot, err := r.kotsAPI.CreateVMSnapshot(r.args.vmSnapshotVMID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "create vm snapshot")
	}

	return print.VMSnapshot(r.outputFormat, r.w, snapshot)
}
