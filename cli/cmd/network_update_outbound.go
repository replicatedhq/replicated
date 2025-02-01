package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkUpdateOutbound(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outbound [ID]",
		Short: "Update outbound setting for a test network.",
		Long:  `The 'outbound' command allows you to update the outbound setting of a test network. The outbound setting can be either 'none' or 'any'.`,
		Example: `# Update the outbound setting for a specific network
replicated network update outbound NETWORK_ID --outbound any`,
		RunE:              r.updateNetworkOutbound,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeNetworkIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateNetworkOutbound, "outbound", "", "Update outbound setting (must be 'none' or 'any')")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide")

	cmd.MarkFlagRequired("outbound")

	return cmd
}

func (r *runners) updateNetworkOutbound(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateNetworkIDArg(args); err != nil {
		return errors.Wrap(err, "ensure network id arg")
	}

	if r.args.updateNetworkOutbound != "none" && r.args.updateNetworkOutbound != "any" {
		return errors.New("outbound must be either 'none' or 'any'")
	}

	opts := kotsclient.UpdateNetworkOutboundOpts{
		Outbound: r.args.updateNetworkOutbound,
	}
	network, err := r.kotsAPI.UpdateNetworkOutbound(r.args.updateNetworkID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	}
	if err != nil {
		return errors.Wrap(err, "update network outbound")
	}

	return print.Network(r.outputFormat, r.w, network)
}
