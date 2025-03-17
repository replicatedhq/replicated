package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkUpdatePolicy(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy [ID]",
		Short: "Update network policy setting for a test network.",
		Long:  `The 'policy' command allows you to update the network policy being used by the network.`,
		Example: `# Update the policy setting for a specific network
replicated network update policy NETWORK_ID --policy airgap`,
		RunE:              r.updateNetworkPolicy,
		SilenceUsage:      true,
		ValidArgsFunction: r.completeNetworkIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.updateNetworkPolicy, "policy", "", "Update network policy setting")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide")

	cmd.MarkFlagRequired("policy")

	return cmd
}

func (r *runners) updateNetworkPolicy(cmd *cobra.Command, args []string) error {
	if err := r.ensureUpdateNetworkIDArg(args); err != nil {
		return errors.Wrap(err, "ensure network id arg")
	}

	if r.args.updateNetworkPolicy == "" {
		return errors.New("policy cannot be empty")
	}

	opts := kotsclient.UpdateNetworkPolicyOpts{
		Policy: r.args.updateNetworkPolicy,
	}
	network, err := r.kotsAPI.UpdateNetworkPolicy(r.args.updateNetworkID, opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	}
	if err != nil {
		return errors.Wrap(err, "update network policy")
	}

	return print.Network(r.outputFormat, r.w, network)
}
