package cmd

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyCreate(parent *cobra.Command) *cobra.Command {
	var (
		name           string
		description    string
		definitionFile string
		outputFormat   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an RBAC policy",
		Long: `Create a new RBAC policy from a JSON definition file.

The definition file must be valid JSON in the following format:
  {
    "v1": {
      "name": "My Policy",
      "resources": {
        "allowed": ["**/*"],
        "denied": []
      }
    }
  }

Vendors not on an enterprise plan cannot create policies.`,
		Example: `  # Create a policy from a definition file
  replicated policy create --name "My Policy" --definition policy.json

  # Create a policy with a description
  replicated policy create --name "My Policy" --description "Custom access policy" --definition policy.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.policyCreate(name, description, definitionFile, outputFormat)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&name, "name", "", "Name of the policy")
	cmd.Flags().StringVar(&description, "description", "", "Description of the policy")
	cmd.Flags().StringVar(&definitionFile, "definition", "", "Path to the JSON file containing the policy definition")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("definition")

	return cmd
}

func (r *runners) policyCreate(name, description, definitionFile, outputFormat string) error {
	definition, err := readPolicyDefinition(definitionFile)
	if err != nil {
		return errors.Wrap(err, "read policy definition")
	}

	policy, err := r.kotsAPI.CreatePolicy(name, description, definition)
	if err != nil {
		if errors.Cause(err) == platformclient.ErrForbidden {
			return errors.New("creating policies requires an enterprise plan")
		}
		return errors.Wrap(err, "create policy")
	}

	return print.Policy(outputFormat, r.w, policy)
}

func readPolicyDefinition(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, "read file")
	}

	if !json.Valid(data) {
		return "", errors.New("policy definition file is not valid JSON")
	}

	return string(data), nil
}
