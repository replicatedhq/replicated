package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAPIPost(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post PATH",
		Short: "Make ad-hoc POST API calls to the Replicated API",
		Long: `This is essentially like curl for the Replicated API, but
uses your local credentials and prints the response unmodified.

We recommend piping the output to jq for easier reading.

Pass the PATH of the request as the final argument (including the API version
prefix). Supported versions are /v1, /v2, and /v3. Do not include the host.`,
		Example: `replicated api post /v3/app/2EuFxKLDxKjPNk2jxMTmF6Vxvxu/channel -b '{"name":"marc-waz-here"}'
replicated api post /v1/app -b '{"name":"My App"}'`,
		RunE:         r.apiPost,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.args.apiPostBody, "body", "b", "", "JSON body to send with the request")

	return cmd
}

func (r *runners) apiPost(cmd *cobra.Command, args []string) error {
	return r.doAPIRequest(cmd.Context(), "POST", args[0], r.args.apiPostBody)
}
