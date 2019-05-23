package shipclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListChannels struct {
	Data   *ShipChannelData `json:"data,omitempty"`
	Errors []GraphQLError   `json:"errors,omitempty"`
}

type ShipChannelData struct {
	ShipChannels []*ShipChannel `json:"getAppChannels"`
}

type ShipChannel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	CurrentSequence int64  `json:"currentSequence"`
	CurrentVersion  string `json:"currentVersion"`
}

func (c *GraphQLClient) ListChannels(appID string) ([]types.Channel, error) {
	response := GraphQLResponseListChannels{}

	request := GraphQLRequest{
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

	if err := c.executeRequest(request, &response); err != nil {
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
	response := GraphQLResponseErrorOnly{}

	request := GraphQLRequest{
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

	if err := c.executeRequest(request, &response); err != nil {
		return err
	}

	if len(response.Errors) != 0 {
		return errors.New(response.Errors[0].Message)
	}

	return nil

}

func ArchiveChannel(appID string, channelID string) error {
	return nil
}

func GetChannel(appID string, channelID string) (interface{}, []interface{}, error) {
	return nil, nil, nil
}
