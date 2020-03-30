package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

const kotsListInstallers = `
query allKotsAppInstallers($appId: ID!) {
  allKotsAppInstallers(appId: $appId) {
	appId
	kurlInstallerId
	sequence
	yaml
	created
	channels {
		id      
		name
		currentVersion
		numReleases
	}    
	isInstallerNotEditable  
  }
} `

type GraphQLResponseListInstallers struct {
	Data   *InstallersDataWrapper `json:"data,omitempty"`
	Errors []graphql.GQLError     `json:"errors,omitempty"`
}

type InstallersDataWrapper struct {
	Installers []types.InstallerSpec `json:"allKotsAppInstallers"`
}

func (c *GraphQLClient) ListInstallers(appID string) ([]types.InstallerSpec, error) {
	response := GraphQLResponseListInstallers{}

	request := graphql.Request{
		Query: kotsListInstallers,

		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute gql request")
	}

	return response.Data.Installers, nil
}

const kotsCreateInstaller = `
mutation createKotsAppInstaller($appId: ID!, $yaml: String!) {
	createKotsAppInstaller(appId: $appId, yaml: $yaml) {
		appId
		kurlInstallerId
		sequence
		created
	}
}`

type GraphQLResponseCreateInstaller struct {
	Data   *CreateInstallerDataWrapper `json:"data,omitempty"`
	Errors []graphql.GQLError          `json:"errors,omitempty"`
}

type CreateInstallerDataWrapper struct {
	Installer *types.InstallerSpec `json:"createKotsAppInstaller"`
}

func (c *GraphQLClient) CreateInstaller(appId string, yaml string) (*types.InstallerSpec, error) {

	installer, err := c.CreateVendorInstaller(appId, yaml)
	if err != nil {
		return nil, errors.Wrap(err, "create vendor installer")
	}

	return installer, nil
}

func (c *GraphQLClient) CreateVendorInstaller(appID string, yaml string) (*types.InstallerSpec, error) {
	response := GraphQLResponseCreateInstaller{}

	request := graphql.Request{
		Query: kotsCreateInstaller,

		Variables: map[string]interface{}{
			"appId":           appID,
			"yaml":            yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute gql request")
	}

	return response.Data.Installer, nil
}

const kotsPromoteInstaller = `
mutation promoteKotsInstaller($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!) {
	promoteKotsInstaller(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel) {
		kurlInstallerId
    }
}`

func (c *GraphQLClient) PromoteInstaller(appID string, sequence int64, channelID string, versionLabel string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: kotsPromoteInstaller,

		Variables: map[string]interface{}{
			"appId":        appID,
			"sequence":     sequence,
			"channelIds":   []string{channelID},
			"versionLabel": versionLabel,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return errors.Wrap(err, "execute gql request")
	}

	return nil

}
