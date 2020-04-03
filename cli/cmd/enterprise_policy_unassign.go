package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyUnassign(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "unassign",
		SilenceUsage: true,
		Short:        "Unassigns a policy from a channel",
		Long: `Remove a new policy from a channel.

  Example:
  replicated enteprise policy unassign --policy-id 123 --channel-id abc`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterprisePolicyUnassignPolicyID, "policy-id", "", "The id of the policy to unassign")
	cmd.Flags().StringVar(&r.args.enterprisePolicyUnassignChannelID, "channel-id", "", "The id of channel")

	cmd.RunE = r.enterprisePolicyUnassign
}

func (r *runners) enterprisePolicyUnassign(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.UnassignPolicy(r.args.enterprisePolicyUnassignPolicyID, r.args.enterprisePolicyUnassignChannelID)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Policy successfully unassigned\n")
	r.w.Flush()

	return nil
}
