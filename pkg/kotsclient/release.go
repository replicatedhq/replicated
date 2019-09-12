package kotsclient

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
)

type GraphQLResponseKotsCreateRelease struct {
	Data   KotsCreateReleaseData `json:"data,omitempty"`
	Errors []graphql.GQLError    `json:"errors,omitempty"`
}

type KotsCreateReleaseData struct {
	KotsReleaseData KotsReleaseSequence `json:"createKotsRelease"`
}

type KotsReleaseSequence struct {
	Sequence int64 `json:"sequence"`
}

type GraphQLResponseKotsUpdateRelease struct {
	Data   KotsUpdateReleaseData `json:"data,omitempty"`
	Errors []graphql.GQLError    `json:"errors,omitempty"`
}

type KotsUpdateReleaseData struct {
	KotsReleaseData KotsReleaseSequence `json:"createKotsRelease"`
}

type GraphQLResponseListReleases struct {
	Data   *KotsReleasesData  `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type KotsReleasesData struct {
	KotsReleases []*KotsRelease `json:"allKotsReleases"`
}

type KotsRelease struct {
	ID           string         `json:"id"`
	Sequence     int64          `json:"sequence"`
	CreatedAt    string         `json:"created"`
	ReleaseNotes string         `json:"releaseNotes"`
	Channels     []*KotsChannel `json:"channels"`
}

type GraphQLResponseUpdateKotsRelease struct {
	Data *KotsReleaseUpdateData `json:"data,omitempty"`
}

type KotsReleaseUpdateData struct {
	UpdateKotsRelease *UpdateKotsRelease `json:"updateKotsRelease"`
}

type UpdateKotsRelease struct {
	ID     string `json:"id"`
	Config string `json:"spec,omitempty"`
}

func (c *GraphQLClient) CreateRelease(appID string, multiyaml string) (*types.ReleaseInfo, error) {
	response := GraphQLResponseKotsCreateRelease{}

	request := graphql.Request{
		Query: `
		mutation createKotsRelease($appId: ID!, $spec: String!) {
			createKotsRelease(appId: $appId, spec: $spec) {
				sequence
			}
		}`,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  multiyaml,
		},
	}
	fmt.Println("multiyaml:", multiyaml)

	if err := c.ExecuteRequest(request, &response); err != nil {
		fmt.Println("anything?", response.Data.KotsReleaseData.Sequence)
		return nil, err
	}

	releaseInfo := types.ReleaseInfo{
		AppID:    appID,
		Sequence: response.Data.KotsReleaseData.Sequence,
	}

	return &releaseInfo, nil
}

var updateKotsRelease = `
  mutation updateKotsRelease($appId: ID!, $spec: String!, $sequence: Int) {
    updateKotsRelease(appId: $appId, spec: $spec, sequence: $sequence) {
      sequence
    }
  }
`

func (c *GraphQLClient) UpdateRelease(appID string, sequence int64, multiyaml string) error {
	response := GraphQLResponseUpdateKotsRelease{}

	request := graphql.Request{
		Query: `
		mutation updateKotsRelease($appId: ID!, $spec: String!, $sequence: Int) {
		  updateKotsRelease(appId: $appId, spec: $spec, sequence: $sequence) {
			sequence
		  }
		}
	  `,

		Variables: map[string]interface{}{
			"appId":    appID,
			"spec":     multiyaml,
			"sequence": sequence,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil
}

var allKotsReleases = `
  query allKotsReleases($appId: ID!, $pageSize: Int, $currentPage: Int) {
    allKotsReleases(appId: $appId, pageSize: $pageSize, currentPage: $currentPage) {
      sequence
      spec
      created
      releaseNotes
      channels {
        id
        name
        currentVersion
        numReleases
      }
      isReleaseNotEditable
    }
  }
`

func (c *GraphQLClient) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	response := GraphQLResponseListReleases{}

	request := graphql.Request{
		Query: allKotsReleases,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	releaseInfos := make([]types.ReleaseInfo, 0, 0)
	for _, kotsRelease := range response.Data.KotsReleases {
		activeChannels := make([]types.Channel, 0, 0)
		createdAt, err := util.ParseTime(kotsRelease.CreatedAt)
		if err != nil {
			return nil, err
		}

		for _, kotsReleaseChannel := range kotsRelease.Channels {
			activeChannel := types.Channel{
				ID:   kotsReleaseChannel.ID,
				Name: kotsReleaseChannel.Name,
			}
			activeChannels = append(activeChannels, activeChannel)

		}
		releaseInfo := types.ReleaseInfo{
			AppID:          appID,
			CreatedAt:      createdAt.In(location),
			EditedAt:       time.Now(),
			Editable:       false,
			Sequence:       kotsRelease.Sequence,
			Version:        "ba",
			ActiveChannels: activeChannels,
		}

		releaseInfos = append(releaseInfos, releaseInfo)
	}

	return releaseInfos, nil
}

var promoteKotsRelease = `
mutation promoteKotsRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $releaseNotes: String) {
    promoteKotsRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, releaseNotes: $releaseNotes) {
      sequence
    }
  }
`

func (c *GraphQLClient) PromoteRelease(appID string, sequence int64, label string, notes string, channelIDs ...string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: promoteKotsRelease,
		Variables: map[string]interface{}{
			"appId":              appID,
			"sequence":           sequence,
			"versionLabel":       label,
			"releaseNotes":       notes,
			"troubleshootSpecId": "",
			"channelIds":         channelIDs,
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
