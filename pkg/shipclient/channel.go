package shipclient

import (
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListChannels struct {
	Data   *ShipChannelData   `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type GraphQLResponseGetChannel struct {
	Data   *ShipGetChannelData `json:"data,omitempty"`
	Errors []graphql.GQLError  `json:"errors,omitempty"`
}

type ShipGetChannelData struct {
	ShipChannel *ShipChannel `json:"getChannel"`
}

type ShipChannelData struct {
	ShipChannels []*ShipChannel `json:"getAppChannels"`
}

type ShipChannel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CurrentSequence int64  `json:"currentSequence"`
	CurrentVersion  string `json:"currentVersion"`
}

func (c *GraphQLClient) ListChannels(appID string) ([]types.Channel, error) {
	response := GraphQLResponseListChannels{}

	request := graphql.Request{
		Query: `
query getAppChannels($appId: ID!) {
  getAppChannels(appId: $appId) {
    id,
    appId,
    name,
    currentVersion,
    currentReleaseDate,
    currentSpec,
    releaseId,
    numReleases,
    description,
    channelIcon
    created,
    updated,
    isDefault,
    isArchived,
    adoptionRate {
      releaseId
      semver
      count
      percent
      totalOnChannel
    }
    releases {
      id
      semver
    }
  }
}`,

		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	channels := make([]types.Channel, 0, 0)
	for _, shipChannel := range response.Data.ShipChannels {
		channel := types.Channel{
			ID:              shipChannel.ID,
			Name:            shipChannel.Name,
			Description:     shipChannel.Description,
			ReleaseSequence: shipChannel.CurrentSequence,
			ReleaseLabel:    shipChannel.CurrentVersion,
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (c *GraphQLClient) CreateChannel(appID string, name string, description string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: `
mutation createChannel($appId: String!, $channelName: String!, $description: String) {
  createChannel(appId: $appId, channelName: $channelName, description: $description) {
    id,
    name,
    description,
    channelIcon,
    currentVersion,
    currentReleaseDate,
    numReleases,
    created,
    updated,
    isDefault,
    isArchived
  }
}`,
		Variables: map[string]interface{}{
			"appId":       appID,
			"channelName": name,
			"description": description,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil

}

func ArchiveChannel(appID string, channelID string) error {
	return nil
}

var getShipChannel = `
  query getChannel($channelId: ID!) {
    getChannel(channelId: $channelId) {
      id
      appId
      name
      description
      channelIcon
      currentVersion
      currentReleaseDate
      installInstructions
      currentSpec
      numReleases
      adoptionRate {
        releaseId
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
        id
        semver
        spec
        releaseNotes
        created
        updated
        sequence
      }
    }
  }
`

func (c *GraphQLClient) GetChannel(appID string, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error) {
	response := GraphQLResponseGetChannel{}

	request := graphql.Request{
		Query: getShipChannel,
		Variables: map[string]interface{}{
			"appID":     appID,
			"channelId": channelID,
		},
	}
	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, nil, err
	}

	channelDetail := channels.AppChannel{
		Id:           response.Data.ShipChannel.ID,
		Name:         response.Data.ShipChannel.Name,
		Description:  response.Data.ShipChannel.Description,
		ReleaseLabel: response.Data.ShipChannel.CurrentVersion,
	}
	return &channelDetail, nil, nil
}
