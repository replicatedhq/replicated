package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListChannels struct {
	Data   *KotsChannelData   `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type KotsChannelData struct {
	KotsChannels []*KotsChannel `json:"getKotsAppChannels"`
}

type KotsChannel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	CurrentSequence int64  `json:"currentSequence"`
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
		currentReleaseDate
		currentSpec
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
			sequence
			semver
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
			Description:     kotsChannel.Description,
			ReleaseSequence: kotsChannel.CurrentSequence,
			ReleaseLabel:    kotsChannel.CurrentVersion,
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func (c *GraphQLClient) CreateChannel(appID string, name string, description string) error {
	response := graphql.ResponseErrorOnly{}

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
