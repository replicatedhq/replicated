package cmd

import (
	"fmt"
	"strings"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitAPIPut(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put",
		Short: "Make ad-hoc PUT API calls to the Replicated API",
		Long: `This is essentially like curl for the Replicated API, but
uses your local credentials and prints the response unmodified.

We recommend piping the output to jq for easier reading.

Pass the PATH of the request as the final argument. Do not include the host or version.

Example:
  replicated api put /v3/app/2EuFxKLDxKjPNk2jxMTmF6Vxvxu/channel/2QLPm10JPkta7jO3Z3Mk4aXTPyZ -b '{"name":"marc-waz-here2"}'
  
`,
		RunE:         r.apiPut,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVarP(&r.args.apiPutBody, "body", "b", "", "JSON body to send with the request")

	return cmd
}

func (r *runners) apiPut(cmd *cobra.Command, args []string) error {
	path := args[0]

	if !strings.HasPrefix(args[0], "/") {
		path = fmt.Sprintf("/%s", args[0])
	}
	pathParts := strings.Split(path, "/")
	// remove any empty parts
	for i := len(pathParts) - 1; i >= 0; i-- {
		if pathParts[i] == "" {
			pathParts = append(pathParts[:i], pathParts[i+1:]...)
		}
	}

	// v1 and v2 paths use platform client, v3 uses kots client
	// split the path on the first slash to determine which client to use
	if pathParts[0] == "v1" {

	} else if pathParts[0] == "v3" {
		kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}
		response, err := kotsRestClient.Put(path, r.args.apiPutBody)
		if err != nil {
			return err
		}

		fmt.Printf("%s", response)
	}

	return nil
}
