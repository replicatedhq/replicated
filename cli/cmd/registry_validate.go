package cmd

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryTest(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "test HOSTNAME",
		Short:        "test registry",
		Long:         `test registry`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.testRegistryImage, "image", "", "The image to test pulling")
	cmd.MarkFlagRequired("image")

	cmd.RunE = r.registryTest
}

func (r *runners) registryTest(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("missing endpoint")
	}
	hostname := args[0]

	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	status, err := kotsRestClient.TestKOTSRegistry(hostname, r.args.testRegistryImage)
	if err != nil {
		return errors.Wrap(err, "test registry")
	}

	if status == http.StatusOK {
		fmt.Println("Registry conection appears ok")
	} else if status == http.StatusNotFound {
		fmt.Println(`Registry connection failed with MANIFEST_UNKNOWN.
Check that the credentials are ok and that the image exists.

If the image exists and the credentials were entered properly, then it's
likely that the credentials do not have access to pull the image specified.

For help, please visit https://docs.replicated.com/external-registry/troubleshooting`)
	} else {
		fmt.Println("Registry connection failed with status", status)
	}

	return nil
}
