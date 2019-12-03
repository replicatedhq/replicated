package kotsclient

import (
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
	Data   *KotsGetChannelData `json:"data,omitempty"`
	Errors []graphql.GQLError  `json:"errors,omitempty"`
}

type KotsGetChannelData struct {
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

func (c *GraphQLClient) ListChannels(appID string) ([]types.Channel, error) {
	response := GraphQLResponseListChannels{}

	request := graphql.Request{
		Query: `
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
	`,

		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
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

func (c *GraphQLClient) CreateChannel(appID string, name string, description string) (string, error) {
	response := GraphQLResponseCreateChannel{}

	request := graphql.Request{
		Query: `
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
	`,
		Variables: map[string]interface{}{
			"appId":       appID,
			"channelName": name,
			"description": description,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return "", err
	}

	return response.Data.KotsChannel.ID, nil

}

func ArchiveChannel(appID string, channelID string) error {
	return nil
}

var getKotsChannel = `
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
