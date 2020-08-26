package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"net/http"
	"os"
)

type GraphQLResponseGetChannel struct {
	Data   *KotsGetChannelData `json:"data,omitempty"`
	Errors []graphql.GQLError  `json:"errors,omitempty"`
}

type GraphQLResponseCreateChannel struct {
	Data   *KotsCreateChannelData `json:"data,omitempty"`
	Errors []graphql.GQLError     `json:"errors,omitempty"`
}

type KotsGetChannelData struct {
	KotsChannel *KotsChannel `json:"getKotsChannel"`
}

type KotsCreateChannelData struct {
	KotsChannel *KotsChannel `json:"createKotsChannel"`
}

type KotsChannel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ChannelSequence int64  `json:"channelSequence"`
	ReleaseSequence int64  `json:"releaseSequence"`
	CurrentVersion  string `json:"currentVersion"`
	ChannelSlug     string `json:"channelSlug"`
}

const embeddedInstallBaseURL = "https://k8s.kurl.sh"

var embeddedInstallOverrideURL = os.Getenv("EMBEDDED_INSTALL_BASE_URL")

// this is not client logic, but sure, let's go with it
func (c *KotsChannel) EmbeddedInstallCommand(appSlug string) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	kurlURL := fmt.Sprintf("%s/%s-%s", kurlBaseURL, appSlug, c.ChannelSlug)
	if c.ChannelSlug == "stable" {
		kurlURL = fmt.Sprintf("%s/%s", kurlBaseURL, appSlug)
	}
	return fmt.Sprintf(`    curl -fsSL %s | sudo bash`, kurlURL)

}

func (c *KotsChannel) EmbeddedAirgapInstallCommand(appSlug string) string {

	kurlBaseURL := embeddedInstallBaseURL
	if embeddedInstallOverrideURL != "" {
		kurlBaseURL = embeddedInstallOverrideURL
	}

	slug := fmt.Sprintf("%s-%s", appSlug, c.ChannelSlug)
	if c.ChannelSlug == "stable" {
		slug = appSlug
	}
	kurlURL := fmt.Sprintf("%s/bundle/%s.tar.gz", kurlBaseURL, slug)

	return fmt.Sprintf(`    curl -fsSL -o %s.tar.gz %s
    tar xvf %s.tar.gz
    sudo bash ./install.sh airgap`, slug, kurlURL, slug)

}

// this is not client logic, but sure, let's go with it
func (c *KotsChannel) ExistingInstallCommand(appSlug string) string {

	slug := appSlug
	if c.ChannelSlug != "stable" {
		slug = fmt.Sprintf("%s/%s", appSlug, c.ChannelSlug)
	}

	return fmt.Sprintf(`    curl -fsSL https://kots.io/install | bash
    kubectl kots install %s`, slug)
}

type ListChannelsResponse struct {
	Channels []*KotsChannel `json:"channels"`
}

func (c *HTTPClient) ListChannels(appID string, appSlug string) ([]types.Channel, error) {
	var response = ListChannelsResponse{}
	url := fmt.Sprintf("/v3/app/%s/channels", appID)
	err := c.DoJSON("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list channels")
	}

	channels := make([]types.Channel, 0, 0)
	for _, kotsChannel := range response.Channels {
		channel := types.Channel{
			ID:              kotsChannel.ID,
			Name:            kotsChannel.Name,
			ReleaseLabel:    kotsChannel.CurrentVersion,
			ReleaseSequence: kotsChannel.ReleaseSequence,
			InstallCommands: &types.InstallCommands{
				Existing: kotsChannel.ExistingInstallCommand(appSlug),
				Embedded: kotsChannel.EmbeddedInstallCommand(appSlug),
				Airgap:   kotsChannel.EmbeddedAirgapInstallCommand(appSlug),
			},
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

const createChannelQuery = `
mutation createKotsChannel($appId: String!, $channelName: String!, $description: String) {
	createKotsChannel(appId: $appId, channelName: $channelName, description: $description) {
	id
	name
	description
	currentVersion
	currentReleaseDate
	numReleases
	created
	updated
	isDefault
	isArchived
	}
}
`

func (c *GraphQLClient) CreateChannel(appID string, name string, description string) (*types.Channel, error) {
	response := GraphQLResponseCreateChannel{}

	request := graphql.Request{
		Query: createChannelQuery,
		Variables: map[string]interface{}{
			"appId":       appID,
			"channelName": name,
			"description": description,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return &types.Channel{
		ID:              response.Data.KotsChannel.ID,
		Name:            response.Data.KotsChannel.Name,
		Description:     response.Data.KotsChannel.Description,
		ReleaseSequence: response.Data.KotsChannel.ReleaseSequence,
		ReleaseLabel:    response.Data.KotsChannel.CurrentVersion,
	}, nil

}

const getKotsChannel = `
query getKotsChannel($channelId: ID!) {
  getKotsChannel(channelId: $channelId) {
    id
    appId
    name
    description
    channelIcon
    channelSequence
    releaseSequence
    currentVersion
    currentReleaseDate
    installInstructions
    numReleases
    adoptionRate {
      releaseSequence
      semver
      count
      percent
      totalOnChannel
    }
    customers {
      id
      name
      avatar
      actions {
	shipApplyDocker
      }
      installationId
      shipInstallStatus {
	status
	updatedAt
      }
    }
    githubRef {
      owner
      repoFullName
      branch
      path
    }
    extraLintRules
    created
    updated
    isDefault
    isArchived
    releases {
      semver
      releaseNotes
      created
      updated
      releasedAt
      sequence
      channelSequence
      airgapBuildStatus
    }
  }
}
`

func (c *GraphQLClient) GetChannel(appID string, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {
	response := GraphQLResponseGetChannel{}

	request := graphql.Request{
		Query: getKotsChannel,
		Variables: map[string]interface{}{
			"appID":     appID,
			"channelId": channelID,
		},
	}
	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, nil, err
	}

	channelDetail := channels.AppChannel{
		Id:              response.Data.KotsChannel.ID,
		Name:            response.Data.KotsChannel.Name,
		Description:     response.Data.KotsChannel.Description,
		ReleaseLabel:    response.Data.KotsChannel.CurrentVersion,
		ReleaseSequence: response.Data.KotsChannel.ReleaseSequence,
	}
	return &channelDetail, nil, nil
}

