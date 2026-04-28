package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyList(parent *cobra.Command) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List RBAC policies",
		Long:    "List all RBAC policies for your team.",
		Example: `  # List all policies
  replicated policy ls

  # List policies in JSON format
  replicated policy ls --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.policyList(outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) policyList(outputFormat string) error {
	policies, err := r.kotsAPI.ListPolicies()
	if err != nil {
		return errors.Wrap(err, "list policies")
	}

	return print.Policies(outputFormat, r.w, policies)
}
