package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm [SNAPSHOT_ID]",
		Aliases: []string{"delete"},
		Short:   "Remove a VM snapshot.",
		Long: `The 'vm snapshot rm' command removes a snapshot from a VM. You must provide the VM ID and either a snapshot ID or the '--all' flag to remove all snapshots for the VM.

After removal, the snapshot files will be cleaned up from the host disk.`,
		Example: `# Remove a snapshot
replicated vm snapshot rm --vm-id VM_ID SNAPSHOT_ID

# Remove all snapshots for a VM
replicated vm snapshot rm --vm-id VM_ID --all`,
		RunE: r.vmSnapshotRemove,
		Args: cobra.MaximumNArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotVMID, "vm-id", "", "The ID of the VM that owns the snapshot (required)")
	err := cmd.MarkFlagRequired("vm-id")
	if err != nil {
		panic(err)
	}
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeVMIDs)
	cmd.Flags().BoolVar(&r.args.vmSnapshotRmAll, "all", false, "Remove all snapshots for the VM")
	cmd.ValidArgsFunction = r.completeVMSnapshotIDs

	return cmd
}

func (r *runners) vmSnapshotRemove(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.vmSnapshotRmAll {
		return errors.New("SNAPSHOT_ID or --all required")
	}
	if len(args) > 0 && r.args.vmSnapshotRmAll {
		return errors.New("cannot specify SNAPSHOT_ID and --all")
	}

	if r.args.vmSnapshotRmAll {
		snapshots, err := r.kotsAPI.ListVMSnapshots(r.args.vmSnapshotVMID)
		if err != nil {
			return errors.Wrap(err, "list vm snapshots")
		}
		for _, s := range snapshots {
			if err := r.kotsAPI.DeleteVMSnapshot(r.args.vmSnapshotVMID, s.ID); err != nil {
				return errors.Wrapf(err, "remove vm snapshot %s", s.ID)
			}
			fmt.Fprintf(r.w, "Snapshot %s has been removed\n", s.ID)
		}
		return r.w.Flush()
	}

	snapshotID := args[0]
	if err := r.kotsAPI.DeleteVMSnapshot(r.args.vmSnapshotVMID, snapshotID); err != nil {
		return errors.Wrap(err, "remove vm snapshot")
	}
	fmt.Fprintf(r.w, "Snapshot %s has been removed\n", snapshotID)
	return r.w.Flush()
}
