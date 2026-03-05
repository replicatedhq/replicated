package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm SNAPSHOT_ID",
		Aliases: []string{"delete"},
		Short:   "Remove a VM snapshot.",
		Long: `The 'vm snapshot rm' command removes a snapshot from a VM. You must provide both the VM ID and the snapshot ID.

After removal, the snapshot files will be cleaned up from the host disk.`,
		Example: `# Remove a snapshot
replicated vm snapshot rm --vm-id VM_ID SNAPSHOT_ID`,
		RunE: r.vmSnapshotRemove,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotVMID, "vm-id", "", "The ID of the VM that owns the snapshot (required)")
	err := cmd.MarkFlagRequired("vm-id")
	if err != nil {
		panic(err)
	}
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeVMIDs)
	cmd.ValidArgsFunction = r.completeVMSnapshotIDs

	return cmd
}

func (r *runners) vmSnapshotRemove(_ *cobra.Command, args []string) error {
	snapshotID := args[0]

	err := r.kotsAPI.DeleteVMSnapshot(r.args.vmSnapshotVMID, snapshotID)
	if err != nil {
		return errors.Wrap(err, "remove vm snapshot")
	}

	fmt.Fprintf(r.w, "Snapshot %s has been removed\n", snapshotID)
	return r.w.Flush()
}
