package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyUpdate(parent *cobra.Command) *cobra.Command {
	var (
		newName        string
		description    string
		definitionFile string
		outputFormat   string
	)

	cmd := &cobra.Command{
		Use:   "update NAME_OR_ID",
		Short: "Update an RBAC policy",
		Long: `Update an existing RBAC policy.

At least one of --name, --description, or --definition must be provided.
The Admin, Read Only, Sales, and Support policies cannot be updated.
Vendors not on an enterprise plan cannot update policies.`,
		Example: `  # Update a policy's definition from a file
  replicated policy update "My Policy" --definition updated-policy.json

  # Rename a policy
  replicated policy update "My Policy" --name "New Policy Name"

  # Update a policy's description and definition
  replicated policy update "My Policy" --description "Updated description" --definition policy.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.policyUpdate(cmd, args[0], newName, description, definitionFile, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&newName, "name", "", "New name for the policy")
	cmd.Flags().StringVar(&description, "description", "", "New description for the policy")
	cmd.Flags().StringVar(&definitionFile, "definition", "", "Path to the JSON file containing the updated policy definition")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	return cmd
}

func (r *runners) policyUpdate(cmd *cobra.Command, nameOrID, newName, description, definitionFile, outputFormat string) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") && !cmd.Flags().Changed("definition") {
		return errors.New("at least one of --name, --description, or --definition must be specified")
	}

	existing, err := r.kotsAPI.GetPolicyByNameOrID(nameOrID)
	if err != nil {
		return errors.Wrap(err, "get policy")
	}

	if existing.ReadOnly {
		return fmt.Errorf("policy %q is read-only and cannot be updated", existing.Name)
	}

	name := existing.Name
	if cmd.Flags().Changed("name") {
		name = newName
	}

	desc := existing.Description
	if cmd.Flags().Changed("description") {
		desc = description
	}

	definition := existing.Definition
	if cmd.Flags().Changed("definition") {
		definition, err = readPolicyDefinition(definitionFile)
		if err != nil {
			return errors.Wrap(err, "read policy definition")
		}
	}

	policy, err := r.kotsAPI.UpdatePolicy(existing.ID, name, desc, definition)
	if err != nil {
		if errors.Cause(err) == platformclient.ErrForbidden {
			return errors.New("updating policies requires an enterprise plan")
		}
		return errors.Wrap(err, "update policy")
	}

	return print.Policy(outputFormat, r.w, policy)
}
