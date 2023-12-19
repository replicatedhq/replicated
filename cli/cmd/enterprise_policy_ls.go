package cmd

import (
	"fmt"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterprisePolicyLS(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Short:        "lists enterprise policies",
		Long:         `lists all policies that have been created`,
		RunE:         r.enterprisePolicyList,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) enterprisePolicyList(cmd *cobra.Command, args []string) error {
	policies, err := r.enterpriseClient.ListPolicies()
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		fmt.Fprintf(r.w, "No policies found. Create one with \"replicated enterprise policy create\"\n")
		r.w.Flush()
		return nil
	}

	print.EnterprisePolicies(r.outputFormat, r.w, policies)
	return nil
}
