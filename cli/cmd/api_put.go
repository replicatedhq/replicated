package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitAPIPut(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put PATH",
		Short: "Make ad-hoc PUT API calls to the Replicated API",
		Long: `This is essentially like curl for the Replicated API, but
uses your local credentials and prints the response unmodified.

We recommend piping the output to jq for easier reading.

Pass the PATH of the request as the final argument (including the API version
prefix). Supported versions are /v1, /v2, and /v3. Do not include the host.`,
		Example: `replicated api put /v3/app/2EuFxKLDxKjPNk2jxMTmF6Vxvxu/channel/2QLPm10JPkta7jO3Z3Mk4aXTPyZ -b '{"name":"marc-waz-here2"}'`,
		RunE:         r.apiPut,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.args.apiPutBody, "body", "b", "", "JSON body to send with the request")

	return cmd
}

func (r *runners) apiPut(cmd *cobra.Command, args []string) error {
	return r.doAPIRequest(cmd.Context(), "PUT", args[0], r.args.apiPutBody)
}
