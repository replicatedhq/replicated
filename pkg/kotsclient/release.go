package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
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
	KotsReleaseData KotsReleaseSequence `json:"updateKotsRelease"`
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

const createReleaseQuery = `
mutation createKotsRelease($appId: ID!, $spec: String) {
	createKotsRelease(appId: $appId, spec: $spec) {
		sequence
	}
}`

func (c *GraphQLClient) CreateRelease(appID string, multiyaml string) (*types.ReleaseInfo, error) {
	response := GraphQLResponseKotsCreateRelease{}

	request := graphql.Request{
		Query: createReleaseQuery,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  multiyaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	releaseInfo := types.ReleaseInfo{
		AppID:    appID,
		Sequence: response.Data.KotsReleaseData.Sequence,
	}

	return &releaseInfo, nil
}

const updateKotsRelease = `
  mutation updateKotsRelease($appId: ID!, $spec: String!, $sequence: Int) {
    updateKotsRelease(appId: $appId, spec: $spec, sequence: $sequence) {
      sequence
    }
  }
`

func (c *GraphQLClient) UpdateRelease(appID string, sequence int64, multiyaml string) error {
	response := GraphQLResponseUpdateKotsRelease{}

	request := graphql.Request{
		Query: updateKotsRelease,

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

func (c *VendorV3Client) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	allReleases := []types.ReleaseInfo{}
	done := false
	page := 0
	for !done {
		resp := types.ListReleasesResponse{}
		path := fmt.Sprintf("/v3/app/%s/releases?currentPage=%d&pageSize=20", appID, page)
		err := c.DoJSON("GET", path, http.StatusOK, nil, &resp)
		if err != nil {
			done = true
			continue
		}
		page += 1
		for _, release := range resp.Releases {
			activeChannels := make([]types.Channel, 0, 0)

			for _, kotsReleaseChannel := range release.Channels {
				if kotsReleaseChannel.IsArchived {
					continue
				}
				activeChannel := types.Channel{
					ID:   kotsReleaseChannel.ID,
					Name: kotsReleaseChannel.Name,
				}
				activeChannels = append(activeChannels, activeChannel)
			}

			newReleaseInfo := types.ReleaseInfo{
				ActiveChannels: activeChannels,
				AppID:          release.AppID,
				CreatedAt:      release.CreatedAt,
				Editable:       !release.IsReleaseNotEditable,
				Sequence:       release.Sequence,
			}
			allReleases = append(allReleases, newReleaseInfo)
		}

		if len(resp.Releases) == 0 {
			done = true
			continue
		}
	}

	return allReleases, nil
}

const promoteKotsRelease = `
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
