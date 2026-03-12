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
		Long: `The 'vm snapshot rm' command removes a snapshot. Provide either a snapshot ID (or short id), '--name' to remove by snapshot name, or the '--all' flag to remove all snapshots.

After removal, the snapshot files will be cleaned up from the host disk.

VM snapshots are currently an alpha feature.`,
		Example: `# Remove a snapshot by ID
replicated vm snapshot rm SNAPSHOT_ID

# Remove a snapshot by name
replicated vm snapshot rm --name "my-snapshot"

# Remove all snapshots
replicated vm snapshot rm --all`,
		RunE: r.vmSnapshotRemove,
		Args: cobra.MaximumNArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.vmSnapshotRmAll, "all", false, "Remove all snapshots")
	cmd.ValidArgsFunction = r.completeVMSnapshotIDsAndNames

	return cmd
}

func (r *runners) vmSnapshotRemove(_ *cobra.Command, args []string) error {
	byID := len(args) > 0
	byName := r.args.vmSnapshotName != ""

	if !byID && !byName && !r.args.vmSnapshotRmAll {
		return errors.New("SNAPSHOT_ID, --name, or --all required")
	}
	if byID && r.args.vmSnapshotRmAll {
		return errors.New("cannot specify SNAPSHOT_ID and --all")
	}
	if byName && r.args.vmSnapshotRmAll {
		return errors.New("cannot specify --name and --all")
	}
	if byID && byName {
		return errors.New("cannot specify SNAPSHOT_ID and --name")
	}

	if r.args.vmSnapshotRmAll {
		snapshots, err := r.kotsAPI.ListVMSnapshots()
		if err != nil {
			return errors.Wrap(err, "list vm snapshots")
		}
		for _, s := range snapshots {
			if err := r.kotsAPI.DeleteVMSnapshot(s.ID); err != nil {
				return errors.Wrapf(err, "remove vm snapshot %s", s.ID[:8])
			}
			fmt.Fprintf(r.w, "Snapshot %s has been removed\n", s.ID[:8])
		}
		return r.w.Flush()
	}

	snapshotIDOrName := args[0]
	if byName {
		snapshotIDOrName = r.args.vmSnapshotName
	}

	if err := r.kotsAPI.DeleteVMSnapshot(snapshotIDOrName); err != nil {
		return errors.Wrap(err, "remove vm snapshot")
	}
	fmt.Fprintf(r.w, "Snapshot %s has been removed\n", snapshotIDOrName)
	return r.w.Flush()
}
