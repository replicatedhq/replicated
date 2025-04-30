package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMUpdateTTL(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ttl [ID_OR_NAME]",
		Short: "Update TTL for a test VM.",
		Long: `The 'ttl' command allows you to update the Time to Live (TTL) for a test VM. This command modifies the lifespan of a running VM by updating its TTL, which is a duration starting from the moment the VM is provisioned.

The TTL specifies how long the VM will run before it is automatically terminated. You can specify a duration up to a maximum of 48 hours. If no TTL is specified, the default TTL is 1 hour.

The command accepts a VM ID or name as an argument and requires the '--ttl' flag to specify the new TTL value.

You can also specify the output format (json, table, wide) using the '--output' flag.`,
		Example: `# Update the TTL of a VM to 2 hours
replicated vm update ttl aaaaa11 --ttl 2h

# Update the TTL of a VM to 30 minutes using VM name
replicated vm update ttl my-test-vm --ttl 30m`,
		RunE:              r.updateVMTTL,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeVMIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateVMTTL, "ttl", "", "Update TTL which starts from the moment the vm is running (duration, max 48h).")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) updateVMTTL(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		vmID, err := r.getVMIDFromArg(args[0])
		if err != nil {
			return errors.Wrap(err, "get vm id from arg")
		}
		r.args.updateVMID = vmID
	} else if err := r.ensureUpdateVMIDArg(args); err != nil {
		return errors.Wrap(err, "ensure vm id arg")
	}

	opts := kotsclient.UpdateVMTTLOpts{
		TTL: r.args.updateVMTTL,
	}
	vm, err := r.kotsAPI.UpdateVMTTL(r.args.updateVMID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "update vm ttl")
	}

	return print.VM(r.outputFormat, r.w, vm)
}
