package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAddOther(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "other",
		Short:        "Add a generic registry",
		Long:         `Add a generic registry using a username/password`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryEndpoint, "endpoint", "", "endpoint for the registry")
	cmd.Flags().StringVar(&r.args.addRegistryUsername, "username", "", "The userame to authenticate to the registry with")
	cmd.Flags().StringVar(&r.args.addRegistryPassword, "password", "", "The password to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryPasswordFromStdIn, "password-stdin", false, "Take the password from stdin")

	cmd.RunE = r.registryAddOther

	return cmd
}

func (r *runners) registryAddOther(cmd *cobra.Command, args []string) error {
	if r.args.addRegistryPasswordFromStdIn {
		var err error
		password, err := r.readPasswordFromStdIn("Password")
		if err != nil {
			return errors.Wrap(err, "read password from stdin")
		}
		r.args.addRegistryPassword = password
	}

	addRegistryRequest, errs := r.validateRegistryAddOther()
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

func (r *runners) validateRegistryAddOther() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "other",
		AuthType: "password",
	}
	errs := []error{}

	if r.args.addRegistryEndpoint == "" {
		errs = append(errs, errors.New("endpoint is required"))
	} else {
		req.Endpoint = r.args.addRegistryEndpoint
	}

	if r.args.addRegistryUsername == "" {
		errs = append(errs, errors.New("username is required"))
	} else {
		req.Username = r.args.addRegistryUsername
	}

	if r.args.addRegistryPassword == "" {
		errs = append(errs, errors.New("password must be specified"))
	} else {
		req.Password = r.args.addRegistryPassword
	}

	return req, errs
}
