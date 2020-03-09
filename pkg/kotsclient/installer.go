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

func (c *GraphQLClient) CreateInstaller(appId string, yaml string) (*types.InstallerSpec, error) {
	return nil, errors.New("not implemented")

}

func (c *GraphQLClient) PromoteInstaller(appID string, sequence int64, channelID string) error {
	return errors.New("not implemented")
}
