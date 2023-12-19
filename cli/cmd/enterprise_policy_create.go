package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "create",
		SilenceUsage: true,
		Short:        "Create a new policy",
		Long: `Create a new policy that can later be assigned to a channel.

  Example:
  replicated enteprise policy create --policy-file myfile.rego --name MyPolicy --description 'A sample policy'`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterprisePolicyCreateName, "name", "", "The name of this policy")
	cmd.Flags().StringVar(&r.args.enterprisePolicyCreateDescription, "description", "", "A longer description of this policy")
	cmd.Flags().StringVar(&r.args.enterprisePolicyCreateFile, "policy-file", "", "The filename containing an OPA policy")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.enterprisePolicyCreate
}

func (r *runners) enterprisePolicyCreate(cmd *cobra.Command, args []string) error {
	b, err := os.ReadFile(r.args.enterprisePolicyCreateFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	policy, err := r.enterpriseClient.CreatePolicy(r.args.enterprisePolicyCreateName, r.args.enterprisePolicyCreateDescription, string(b))
	if err != nil {
		return err
	}

	print.EnterprisePolicy(r.outputFormat, r.w, policy)
	return nil
}
