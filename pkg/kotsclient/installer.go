package kotsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
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
mutation createKotsAppInstaller($appId: ID!, $kurlInstallerId: ID!, $yaml: String!) {
	createKotsAppInstaller(appId: $appId, kurlInstallerId: $kurlInstallerId, yaml: $yaml) {
		appId
		kurlInstallerId
		sequence
		created
	}
}`

const kotsInstallerForSequence = `
query kotsInstallerForSequence($appId: ID!, $sequence: Int) {
	kotsInstallerForSequence(appId: $appId, sequence: $sequence) {
		appId
		sequence
		yaml
		channels {
			id
	    	name
		}
		isInstallerNotEditable
	}
}`

type GraphQLResponseCreateInstaller struct {
	Data   *InstallersDataWrapper `json:"data,omitempty"`
	Errors []graphql.GQLError     `json:"errors,omitempty"`
}

type CreateInstallerDataWrapper struct {
	Installers []types.InstallerSpec `json:"allKotsAppInstallers"`
}

func (c *GraphQLClient) CreateInstaller(appId string, yaml string) (*types.InstallerSpec, error) {

	// post yaml to kurl.sh
	installerURL, err := c.CreateKurldotSHInstaller(yaml)
	if err != nil {
		return nil, errors.Wrap(err, "create kurl installer")
	}

	trimmed := strings.TrimLeft(installerURL, "htps:/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2{
		return nil, errors.Errorf("expected exactly two parts of %q, found %d", trimmed, len(parts))
	}

	installerKurlHash := parts[1]

	// run create installer with kurl hash
	// return installer
	return nil, errors.New("not implemented")

}


func (c *GraphQLClient) CreateKurldotSHInstaller(yaml string) (string, error) {
	bodyReader := strings.NewReader(yaml)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/installer", c.KurlDotSHAddress), bodyReader)
	if err != nil {
		return "", errors.Wrap(err, "create request")
	}

	req.Header.Set("Content-Type", "text/yaml")

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "do request")
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read response body")
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code %d, body was %s", resp.StatusCode, responseBody)
	}

	return string(responseBody), nil
}

func (c *GraphQLClient) PromoteInstaller(appID string, sequence int64, channelID string) error {
	return errors.New("not implemented")
}
