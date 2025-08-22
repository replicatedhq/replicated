package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitRegistryAddGHCR(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "ghcr",
		Short:        "Add a GitHub Container Registry",
		Long:         `Add a GitHub Container Registry using a username and personal access token (PAT)`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryUsername, "username", "", "The username to authenticate to the registry with")
	cmd.Flags().StringVar(&r.args.addRegistryToken, "token", "", "The token to use to auth to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryTokenFromStdIn, "token-stdin", false, "Take the token from stdin")
	cmd.Flags().StringVar(&r.args.addRegistryName, "name", "", "Name for the registry")
	cmd.Flags().StringVar(&r.args.addRegistryAppIds, "app-ids", "", "Comma-separated list of app IDs to scope this registry to")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.RunE = r.registryAddGHCR
}

func (r *runners) registryAddGHCR(cmd *cobra.Command, args []string) error {
	if r.args.addRegistryTokenFromStdIn {
		var err error
		token, err := r.readPasswordFromStdIn("Token")
		if err != nil {
			return errors.Wrap(err, "read token from stdin")
		}
		r.args.addRegistryToken = token
	}

	addRegistryRequest, errs := r.validateRegistryAddGHCR()
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

func (r *runners) validateRegistryAddGHCR() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "ghcr",
		AuthType: "token",
		Endpoint: "ghcr.io",
	}

	// Handle name/slug
	if r.args.addRegistryName != "" {
		req.Slug = r.args.addRegistryName
	} else {
		req.Slug = req.Endpoint
	}

	// Parse app IDs
	if r.args.addRegistryAppIds != "" {
		appIds := strings.Split(r.args.addRegistryAppIds, ",")
		for i, id := range appIds {
			appIds[i] = strings.TrimSpace(id)
		}
		req.AppIds = appIds
	}
	errs := []error{}

	if r.args.addRegistryUsername == "" {
		errs = append(errs, errors.New("username is required"))
	} else {
		req.Username = r.args.addRegistryUsername
	}

	if r.args.addRegistryToken == "" {
		errs = append(errs, errors.New("token is required"))
	} else {
		req.Password = r.args.addRegistryToken
	}

	return req, errs
}
