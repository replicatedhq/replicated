package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

// Google Artifact Registry
func (r *runners) InitRegistryAddGAR(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "gar",
		Short:        "Add a Google Artifact Registry",
		Long:         `Add a Google Artifact Registry using a service account key`,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.addRegistryEndpoint, "endpoint", "", "The GAR endpoint")
	cmd.Flags().StringVar(&r.args.addRegistryAuthType, "authtype", "serviceaccount", "Auth type for the registry")
	cmd.Flags().StringVar(&r.args.addRegistryServiceAccountKey, "serviceaccountkey", "", "The service account key to authenticate to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryServiceAccountKeyFromStdIn, "serviceaccountkey-stdin", false, "Take the service account key from stdin")
	cmd.Flags().StringVar(&r.args.addRegistryToken, "token", "", "The token to use to auth to the registry with")
	cmd.Flags().BoolVar(&r.args.addRegistryTokenFromStdIn, "token-stdin", false, "Take the token from stdin")
	cmd.Flags().StringVar(&r.args.addRegistryName, "name", "", "Name for the registry")
	cmd.Flags().StringVar(&r.args.addRegistryAppIds, "app-ids", "", "Comma-separated list of app IDs to scope this registry to")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")

	cmd.RunE = r.registryAddGAR
}

func (r *runners) registryAddGAR(cmd *cobra.Command, args []string) error {

	if r.args.addRegistryServiceAccountKeyFromStdIn {
		var err error
		serviceAccountKey, err := r.readPasswordFromStdIn("Service Account Key")
		if err != nil {
			return errors.Wrap(err, "read secret service account key from stdin")
		}
		r.args.addRegistryServiceAccountKey = serviceAccountKey
	}

	if r.args.addRegistryTokenFromStdIn {
		var err error
		token, err := r.readPasswordFromStdIn("Token")
		if err != nil {
			return errors.Wrap(err, "read token from stdin")
		}
		r.args.addRegistryToken = token
	}

	addRegistryRequest, errs := r.validateRegistryAddGAR()
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

func (r *runners) validateRegistryAddGAR() (kotsclient.AddKOTSRegistryRequest, []error) {
	req := kotsclient.AddKOTSRegistryRequest{
		Provider: "gar",
		AuthType: "serviceaccount",
		Username: "_json_key",
	}

	// Handle name/slug
	if r.args.addRegistryName != "" {
		req.Slug = r.args.addRegistryName
	} else {
		req.Slug = r.args.addRegistryEndpoint
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

	if r.args.addRegistryEndpoint == "" {
		errs = append(errs, errors.New("endpoint must be specified"))
	} else {
		req.Endpoint = r.args.addRegistryEndpoint
	}

	supportedAuthTypes := []string{"serviceaccount", "token"}
	if !contains(supportedAuthTypes, r.args.addRegistryAuthType) {
		errs = append(errs, errors.New("authtype must be one of: serviceaccount, token"))
	} else {
		req.AuthType = r.args.addRegistryAuthType
	}

	if req.AuthType == "serviceaccount" {
		if r.args.addRegistryServiceAccountKey == "" {
			errs = append(errs, errors.New("serviceaccountkey or serviceaccountkey-stdin must be specified"))
		} else {
			req.Username = "_json_key"
			req.Password = r.args.addRegistryServiceAccountKey
		}

	} else if req.AuthType == "token" {

		if r.args.addRegistryToken == "" {
			errs = append(errs, errors.New("token is required"))
		} else {
			req.Username = "oauth2accesstoken"
			req.Password = r.args.addRegistryToken
		}
	}

	return req, errs
}
