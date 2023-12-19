package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAddGCR(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "gcr",
		Short:        "Add a Google Container Registry",
		Long:         `Add a Google Container Registry using a service account key`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryEndpoint, "endpoint", "", "The GCR endpoint")
	cmd.Flags().StringVar(&r.args.addRegistryServiceAccountKey, "serviceaccountkey", "", "The service account key to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryServiceAccountKeyFromStdIn, "serviceaccountkey-stdin", false, "Take the service account key from stdin")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.registryAddGCR
}

func (r *runners) registryAddGCR(cmd *cobra.Command, args []string) error {
	if r.args.addRegistryServiceAccountKeyFromStdIn {
		var err error
		serviceAccountKey, err := r.readPasswordFromStdIn("Service Account Key")
		if err != nil {
			return errors.Wrap(err, "read secret service account key from stdin")
		}
		r.args.addRegistryServiceAccountKey = serviceAccountKey
	}

	addRegistryRequest, errs := r.validateRegistryAddGCR()
	if len(errs) > 0 {
		joinedErrs := []string{}
		for _, err := range errs {
			joinedErrs = append(joinedErrs, err.Error())
		}

		return errors.New(strings.Join(joinedErrs, ", "))
	}

	addRegistryRequest.SkipValidation = r.args.addRegistrySkipValidation

	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	registry, err := kotsRestClient.AddKOTSRegistry(addRegistryRequest)
	if err != nil {
		return errors.Wrap(err, "add registry")
	}

	registries := []types.Registry{
		*registry,
	}

	return print.Registries(r.outputFormat, r.w, registries)

}

func (r *runners) validateRegistryAddGCR() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "gcr",
		AuthType: "serviceaccount",
		Username: "_json_key",
	}
	errs := []error{}

	if r.args.addRegistryEndpoint == "" {
		errs = append(errs, errors.New("endpoint must be specified"))
	} else {
		req.Endpoint = r.args.addRegistryEndpoint
	}

	if r.args.addRegistryServiceAccountKey == "" {
		errs = append(errs, errors.New("serviceaccountkey or serviceaccountkey-stdin must be specified"))
	} else {
		req.Password = r.args.addRegistryServiceAccountKey
	}

	return req, errs
}
