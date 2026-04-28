package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm NAME_OR_ID",
		Aliases: []string{"delete"},
		Short:   "Remove an RBAC policy",
		Long: `Remove an RBAC policy.

The Admin, Read Only, Sales, and Support policies cannot be removed.
Vendors not on an enterprise plan cannot remove policies.`,
		Example: `  # Remove a policy by name
  replicated policy rm "My Policy"

  # Remove a policy by ID
  replicated policy rm pol_abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.policyRm(args[0])
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) policyRm(nameOrID string) error {
	policy, err := r.kotsAPI.GetPolicyByNameOrID(nameOrID)
	if err != nil {
		return errors.Wrap(err, "get policy")
	}

	if policy.ReadOnly {
		return fmt.Errorf("policy %q is read-only and cannot be removed", policy.Name)
	}

	if err := r.kotsAPI.DeletePolicy(policy.ID); err != nil {
		if errors.Cause(err) == platformclient.ErrForbidden {
			return errors.New("removing policies requires an enterprise plan")
		}
		return errors.Wrap(err, "remove policy")
	}

	fmt.Fprintf(r.w, "Policy %s removed.\n", policy.Name)
	return r.w.Flush()
}
