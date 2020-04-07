package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyRM(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "rm",
		SilenceUsage: true,
		Short:        "Remove a policy",
		Long: `Remove a policy.

  Example:
  replicated enteprise policy rm --id MyPolicyID`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterprisePolicyRmId, "id", "", "The id of the policy to remove")

	cmd.RunE = r.enterprisePolicyRemove
}

func (r *runners) enterprisePolicyRemove(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.RemovePolicy(r.args.enterprisePolicyRmId)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Policy %s successfully removed\n", r.args.enterprisePolicyRmId)
	r.w.Flush()

	return nil
}
