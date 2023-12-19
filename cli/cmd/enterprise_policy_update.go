package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "update",
		SilenceUsage: true,
		Short:        "update an existing policy",
		Long: `Update an existing policy.

  Example:
  replicated enteprise policy update --id MyPolicyID --policy-file myfile.rego --name MyPolicy --description 'A sample policy'`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterprisePolicyUpdateID, "id", "", "The id of the policy to be updated")
	cmd.Flags().StringVar(&r.args.enterprisePolicyUpdateName, "name", "", "The new name for this policy")
	cmd.Flags().StringVar(&r.args.enterprisePolicyUpdateDescription, "description", "", "The new description of this policy")
	cmd.Flags().StringVar(&r.args.enterprisePolicyUpdateFile, "policy-file", "", "The filename containing an OPA policy")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.enterprisePolicyUpdate
}

func (r *runners) enterprisePolicyUpdate(cmd *cobra.Command, args []string) error {
	b, err := os.ReadFile(r.args.enterprisePolicyUpdateFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	policy, err := r.enterpriseClient.UpdatePolicy(r.args.enterprisePolicyUpdateID, r.args.enterprisePolicyUpdateName, r.args.enterprisePolicyUpdateDescription, string(b))
	if err != nil {
		return err
	}

	print.EnterprisePolicy(r.outputFormat, r.w, policy)
	return nil
}
