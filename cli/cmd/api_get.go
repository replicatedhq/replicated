package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAPIGet(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get PATH",
		Short: "Make ad-hoc GET API calls to the Replicated API",
		Long: `This is essentially like curl for the Replicated API, but
uses your local credentials and prints the response unmodified.

We recommend piping the output to jq for easier reading.

Pass the PATH of the request as the final argument (including the API version
prefix). Supported versions are /v1, /v2, and /v3. Do not include the host.`,
		Example: `replicated api get /v3/apps
replicated api get /v1/apps`,
		RunE:         r.apiGet,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) apiGet(cmd *cobra.Command, args []string) error {
	return r.doAPIRequest(cmd.Context(), "GET", args[0], "")
}
