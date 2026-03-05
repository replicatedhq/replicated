package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMSnapshotLs(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List snapshots for a VM.",
		Long: `The 'vm snapshot ls' command lists all snapshots for a specific VM. You must provide the VM ID using the --vm-id flag.

This command is useful for viewing existing snapshots, their status, size, and creation time. The output format can be customized using the --output flag.`,
		Example: `# List snapshots for a VM
replicated vm snapshot ls --vm-id VM_ID

# List snapshots in JSON format
replicated vm snapshot ls --vm-id VM_ID --output json`,
		RunE: r.vmSnapshotList,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.vmSnapshotVMID, "vm-id", "", "The ID of the VM to list snapshots for (required)")
	err := cmd.MarkFlagRequired("vm-id")
	if err != nil {
		panic(err)
	}
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeVMIDs)
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotList(_ *cobra.Command, args []string) error {
	snapshots, err := r.kotsAPI.ListVMSnapshots(r.args.vmSnapshotVMID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list vm snapshots")
	}

	if len(snapshots) == 0 {
		return print.NoVMSnapshots(r.outputFormat, r.w)
	}

	return print.VMSnapshots(r.outputFormat, r.w, snapshots, true)
}
