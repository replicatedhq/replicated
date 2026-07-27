package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAPIPatch(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch PATH",
		Short: "Make ad-hoc PATCH API calls to the Replicated API",
		Long: `This is essentially like curl for the Replicated API, but
uses your local credentials and prints the response unmodified.

We recommend piping the output to jq for easier reading.

Pass the PATH of the request as the final argument (including the API version
prefix). Supported versions are /v1, /v2, and /v3. Do not include the host.`,
		Example: `replicated api patch /v3/customer/2VffY549paATVfHSGpJhjh6Ehpy -b '{"name":"Valuable Customer"}'`,
		RunE:         r.apiPatch,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.args.apiPatchBody, "body", "b", "", "JSON body to send with the request")

	return cmd
}

func (r *runners) apiPatch(cmd *cobra.Command, args []string) error {
	return r.doAPIRequest(cmd.Context(), "PATCH", args[0], r.args.apiPatchBody)
}
