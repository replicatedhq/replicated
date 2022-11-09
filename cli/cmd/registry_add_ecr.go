package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAddECR(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "ecr",
		Short:        "Add an ECR registry",
		Long:         `Add an ECR registry using an Access Key ID and Secret Access Key`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryEndpoint, "endpoint", "", "The ECR endpoint")
	cmd.Flags().StringVar(&r.args.addRegistryAccessKeyID, "accesskeyid", "", "The access key id to authenticate to the registry with")
	cmd.Flags().StringVar(&r.args.addRegistrySecretAccessKey, "secretaccesskey", "", "The secret access key to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistrySecretAccessKeyFromStdIn, "secretaccesskey-stdin", false, "Take the secret access key from stdin")

	cmd.RunE = r.registryAddECR
}

func (r *runners) registryAddECR(cmd *cobra.Command, args []string) error {
	if r.args.addRegistrySecretAccessKeyFromStdIn {
		var err error
		secretAccessKey, err := r.readPasswordFromStdIn("Secret Access Key")
		if err != nil {
			return errors.Wrap(err, "read secret access key from stdin")
		}
		r.args.addRegistrySecretAccessKey = secretAccessKey
	}

	addRegistryRequest, errs := r.validateRegistryAddECR()
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

func (r *runners) validateRegistryAddECR() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "ecr",
		AuthType: "accesskey",
	}
	errs := []error{}

	if r.args.addRegistryEndpoint == "" {
		errs = append(errs, errors.New("endpoint must be specified"))
	} else {
		req.Endpoint = r.args.addRegistryEndpoint
	}

	if r.args.addRegistryAccessKeyID == "" {
		errs = append(errs, errors.New("accesskeyid must be specified"))
	} else {
		req.Username = r.args.addRegistryAccessKeyID
	}

	if r.args.addRegistrySecretAccessKey == "" {
		errs = append(errs, errors.New("secretaccesskey or secretaccesskey-stdin must be specified"))
	} else {
		req.Password = r.args.addRegistrySecretAccessKey
	}

	return req, errs
}
