package cmd

import (
	"encoding/json"
	"io/ioutil"
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
	cmd.Flags().StringVar(&r.args.addRegistryServiceAccountKey, "serviceaccountkey", "", "The service account key to authenticate to the registry with. This is the path to the JSON file.")
	cmd.Flags().BoolVar(&r.args.addRegistryServiceAccountKeyFromStdIn, "serviceaccountkey-stdin", false, "Take the service account key content from stdin")

	cmd.RunE = r.registryAddGCR
}

func (r *runners) registryAddGCR(cmd *cobra.Command, args []string) error {
	if r.args.addRegistryServiceAccountKeyFromStdIn {
		var err error
		serviceAccountKey, err := r.readPasswordFromStdIn("Service Account Key")
		if err != nil {
			return errors.Wrap(err, "read secret service account key from stdin")
		}
		r.args.addRegistryPassword = serviceAccountKey
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

	return print.Registries(r.w, registries)

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

	if r.args.addRegistryServiceAccountKey == "" && r.args.addRegistryPassword == "" {
		errs = append(errs, errors.New("serviceaccountkey or serviceaccountkey-stdin must be specified"))
	} else {
		if r.args.addRegistryServiceAccountKey != "" {
			bytes, err := ioutil.ReadFile(r.args.addRegistryServiceAccountKey)
			if err != nil {
				errs = append(errs, errors.Wrap(err, "read service account key"))
				return req, errs
			}
			if !json.Valid(bytes) {
				errs = append(errs, errors.New("Not valid json key file"))
				return req, errs
			}
			r.args.addRegistryPassword = string(bytes)
		}
		req.Password = r.args.addRegistryPassword
	}

	return req, errs
}
