package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/util"
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

# Restore with an SSH public key
replicated vm snapshot restore --vm-id VM_ID SNAPSHOT_ID --ssh-public-key ~/.ssh/id_ed25519.pub

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
	cmd.RegisterFlagCompletionFunc("vm-id", r.completeVMIDs)
	cmd.ValidArgsFunction = r.completeVMSnapshotIDs
	cmd.Flags().StringArrayVar(&r.args.vmSnapshotPublicKeys, "ssh-public-key", []string{}, "Path to SSH public key file to add to the VM (can be specified multiple times)")
	cmd.Flags().StringVar(&r.args.vmSnapshotTTL, "ttl", "", "VM TTL (duration)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmSnapshotRestore(_ *cobra.Command, args []string) error {
	snapshotID := args[0]

	var publicKeys []string
	for _, keyPath := range r.args.vmSnapshotPublicKeys {
		publicKey, err := util.ReadAndValidatePublicKey(keyPath)
		if err != nil {
			return errors.Wrap(err, "validate public key")
		}
		publicKeys = append(publicKeys, publicKey)
	}

	vm, err := r.kotsAPI.RestoreVMSnapshot(r.args.vmSnapshotVMID, snapshotID, publicKeys, r.args.vmSnapshotTTL)
	if err != nil {
		if strings.Contains(err.Error(), "no public keys available") {
			return errors.New("no public keys available; use --ssh-public-key to provide a key or add SSH keys to your team")
		}
		return errors.Wrap(err, "restore vm snapshot")
	}

	return print.VM(r.outputFormat, r.w, vm)
}
