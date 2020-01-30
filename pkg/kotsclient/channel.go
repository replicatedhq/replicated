package kotsclient

import (
	"fmt"
	"github.com/pkg/errors"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListChannels struct {
	Data   *KotsChannelData   `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

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
type KotsChannelData struct {
	KotsChannels []*KotsChannel `json:"getKotsAppChannels"`
}

type KotsChannel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ChannelSequence int64  `json:"channelSequence"`
	ReleaseSequence int64  `json:"releaseSequence"`
	CurrentVersion  string `json:"currentVersion"`
}

const listChannelsQuery = `
query getKotsAppChannels($appId: ID!) {
	getKotsAppChannels(appId: $appId) {
	id
	appId
	name
	currentVersion
	channelSequence
	releaseSequence
	currentReleaseDate
	numReleases
	description
	channelIcon
	created
	updated
	isDefault
	isArchived
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
		shipInstallStatus {
		status
		}
	}
	githubRef {
		owner
		repoFullName
		branch
		path
	}
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

func (c *HybridClient) ListChannels(appID string) ([]types.Channel, error) {
	response := GraphQLResponseListChannels{}

	request := graphql.Request{
		Query: listChannelsQuery,

		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteGraphQLRequest(request, &response); err != nil {
		return nil, err
	}

	channels := make([]types.Channel, 0, 0)
	for _, kotsChannel := range response.Data.KotsChannels {
		channel := types.Channel{
			ID:              kotsChannel.ID,
			Name:            kotsChannel.Name,
			ReleaseLabel:    kotsChannel.CurrentVersion,
			ReleaseSequence: kotsChannel.ReleaseSequence,
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

func (c *HybridClient) CreateChannel(appID string, name string, description string) (*types.Channel, error) {
	response := GraphQLResponseCreateChannel{}

	request := graphql.Request{
		Query: createChannelQuery,
		Variables: map[string]interface{}{
			"appId":       appID,
			"channelName": name,
			"description": description,
		},
	}

	if err := c.ExecuteGraphQLRequest(request, &response); err != nil {
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

func ArchiveChannel(appID string, channelID string) error {
	return nil
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

func (c *HybridClient) GetChannel(appID string, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {
	response := GraphQLResponseGetChannel{}

	request := graphql.Request{
		Query: getKotsChannel,
		Variables: map[string]interface{}{
			"appID":     appID,
			"channelId": channelID,
		},
	}
	if err := c.ExecuteGraphQLRequest(request, &response); err != nil {
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

func (c *HybridClient) GetChannelByName(appID string, name string, description string, create bool) (*types.Channel, error) {
	allChannels, err := c.ListChannels(appID)
	if err != nil {
		return nil, err
	}

	matchingChannels := make([]*types.Channel, 0)
	for _, channel := range allChannels {
		if channel.ID == name || channel.Name == name {
			matchingChannels = append(matchingChannels, &types.Channel{
				ID:              channel.ID,
				Name:            channel.Name,
				Description:     channel.Description,
				ReleaseSequence: channel.ReleaseSequence,
				ReleaseLabel:    channel.ReleaseLabel,
			})
		}
	}

	if len(matchingChannels) == 0 {
		if create {
			channel, err := c.CreateChannel(appID, name, description)
			if err != nil {
				return nil, errors.Wrapf(err, "create channel %q ", name)
			}
			return channel, nil
		}

		return nil, fmt.Errorf("could not find channel %q", name)
	}

	if len(matchingChannels) > 1 {
		return nil, fmt.Errorf("channel %q is ambiguous, please use channel ID", name)
	}
	return matchingChannels[0], nil
}
