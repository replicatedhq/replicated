package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyAssign(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "assign",
		SilenceUsage: true,
		Short:        "Assigns a policy to a channel",
		Long: `Assigns a policy to a channel.

  Example:
  replicated enteprise policy assign --policy-id 123 --channel-id abc`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterprisePolicyAssignPolicyID, "policy-id", "", "The id of the policy to assign")
	cmd.Flags().StringVar(&r.args.enterprisePolicyAssignChannelID, "channel-id", "", "The id of channel")

	cmd.RunE = r.enterprisePolicyAssign
}

func (r *runners) enterprisePolicyAssign(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AssignPolicy(r.args.enterprisePolicyAssignPolicyID, r.args.enterprisePolicyAssignChannelID)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Policy successfully assigned\n")
	r.w.Flush()

	return nil
}
