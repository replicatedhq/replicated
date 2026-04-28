package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitPolicyGet(parent *cobra.Command) *cobra.Command {
	var (
		outputFormat string
		outputFile   string
	)

	cmd := &cobra.Command{
		Use:   "get NAME_OR_ID",
		Short: "Get an RBAC policy",
		Long:  "Display details for an RBAC policy. Use --output-file to save the policy definition to a JSON file.",
		Example: `  # Get a policy by name
  replicated policy get "My Policy"

  # Get a policy and save its definition to a file
  replicated policy get "My Policy" --output-file policy.json

  # Get a policy in JSON format
  replicated policy get "My Policy" --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.policyGet(args[0], outputFormat, outputFile)
		},
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "If set, saves the policy definition to the specified file")

	return cmd
}

func (r *runners) policyGet(nameOrID, outputFormat, outputFile string) error {
	policy, err := r.kotsAPI.GetPolicyByNameOrID(nameOrID)
	if err != nil {
		return errors.Wrap(err, "get policy")
	}

	if outputFile != "" {
		b, err := json.MarshalIndent(json.RawMessage(policy.Definition), "", "  ")
		if err != nil {
			return errors.Wrap(err, "marshal policy definition")
		}
		if err := os.WriteFile(outputFile, append(b, '\n'), 0644); err != nil {
			return errors.Wrap(err, "write policy file")
		}
		fmt.Fprintf(r.w, "Policy definition saved to %s\n", outputFile)
		return r.w.Flush()
	}

	return print.Policy(outputFormat, r.w, policy)
}
