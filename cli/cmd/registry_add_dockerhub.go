package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAddDockerHub(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dockerhub",
		Short:        "Add a DockerHub registry",
		Long:         `Add a DockerHub registry using a username/password or an account token`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryAuthType, "authtype", "password", "Auth type for the registry")
	cmd.Flags().StringVar(&r.args.addRegistryUsername, "username", "", "The userame to authenticate to the registry with")
	cmd.Flags().StringVar(&r.args.addRegistryPassword, "password", "", "The password to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryPasswordFromStdIn, "password-stdin", false, "Take the password from stdin")
	cmd.Flags().StringVar(&r.args.addRegistryToken, "token", "", "The token to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryTokenFromStdIn, "token-stdin", false, "Take the token from stdin")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.registryAddDockerHub

	return cmd
}

func (r *runners) registryAddDockerHub(cmd *cobra.Command, args []string) error {
	if r.args.addRegistryPasswordFromStdIn {
		var err error
		password, err := r.readPasswordFromStdIn("Password")
		if err != nil {
			return errors.Wrap(err, "read password from stdin")
		}
		r.args.addRegistryPassword = password
	}
	if r.args.addRegistryTokenFromStdIn {
		var err error
		password, err := r.readPasswordFromStdIn("Token")
		if err != nil {
			return errors.Wrap(err, "read token from stdin")
		}
		r.args.addRegistryPassword = password
	}

	addRegistryRequest, errs := r.validateRegistryAddDockerHub()
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

func (r *runners) validateRegistryAddDockerHub() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "dockerhub",
		Endpoint: "index.docker.io",
	}
	errs := []error{}

	supportedAuthTypes := []string{"password", "token"}
	if !contains(supportedAuthTypes, r.args.addRegistryAuthType) {
		errs = append(errs, errors.New("authtype must be one of: password, token"))
	} else {
		req.AuthType = r.args.addRegistryAuthType
	}

	if r.args.addRegistryUsername == "" {
		errs = append(errs, errors.New("username is required"))
	} else {
		req.Username = r.args.addRegistryUsername
	}

	if r.args.addRegistryPassword == "" {
		if r.args.addRegistryAuthType == "password" {
			errs = append(errs, errors.New("password must be specified"))
		} else {
			errs = append(errs, errors.New("token or token-stdin must be specified"))
		}
	} else {
		req.Password = r.args.addRegistryPassword
	}

	return req, errs
}
