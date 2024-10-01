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
		Use:               "ttl [ID]",
		Short:             "Update TTL for a test vm",
		Long:              `Update TTL for a test vm`,
		RunE:              r.updateVMTTL,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateClusterTTL, "ttl", "", "Update TTL which starts from the moment the vm is running (duration, max 48h).")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	cmd.MarkFlagRequired("ttl")

	return cmd
}

func (r *runners) updateVMTTL(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateVMIDArg(args); err != nil {
		return errors.Wrap(err, "ensure vm id arg")
	}

	opts := kotsclient.UpdateVMTTLOpts{
		TTL: r.args.updateClusterTTL,
	}
	vm, err := r.kotsAPI.UpdateVMTTL(r.args.updateClusterID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "update vm ttl")
	}

	return print.VM(r.outputFormat, r.w, vm)
}
