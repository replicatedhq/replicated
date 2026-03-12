package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotUpdate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [SNAPSHOT_ID]",
		Short: "Update VM snapshot TTL.",
		Long: `The 'vm snapshot update' command updates the Time to Live (TTL) for a snapshot. You can specify the snapshot by ID (or short id) or by name using the '--name' flag.

TTL is bounded by the team max TTL (same as VMs).

VM snapshots are currently an alpha feature.`,
		Example: `# Update snapshot TTL by ID
replicated vm snapshot update SNAPSHOT_ID --ttl 48h

# Update snapshot TTL by name
replicated vm snapshot update --name "my-snapshot" --ttl 48h`,
		RunE:              r.vmSnapshotUpdate,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeVMSnapshotIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotUpdateTTL, "ttl", "", "New TTL for the snapshot (e.g. \"48h\", \"24h\"). Bounded by team max TTL.")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")
	cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) vmSnapshotUpdate(_ *cobra.Command, args []string) error {
	snapshotIDOrName, err := r.ensureSnapshotIDArg(args)
	if err != nil {
		return err
	}

	snapshot, err := r.kotsAPI.UpdateVMSnapshot(snapshotIDOrName, r.args.vmSnapshotUpdateTTL)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	}
	if err != nil {
		return errors.Wrap(err, "update vm snapshot")
	}

	return print.VMSnapshot(r.outputFormat, r.w, snapshot)
}
