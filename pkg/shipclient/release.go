package shipclient

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
)

type GraphQLResponseListReleases struct {
	Data   *ShipReleasesData  `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type ShipReleasesData struct {
	ShipReleases []*ShipRelease `json:"allReleases"`
}

type ShipRelease struct {
	ID           string         `json:"id"`
	Sequence     int64          `json:"sequence"`
	CreatedAt    string         `json:"created"`
	ReleaseNotes string         `json:"releaseNotes"`
	Channels     []*ShipChannel `json:"channels"`
}

type GraphQLResponseFinalizeRelease struct {
	Data   *ShipFinalizeCreateData `json:"data,omitempty"`
	Errors []graphql.GQLError      `json:"errors,omitempty"`
}

type ShipFinalizeCreateData struct {
	ShipRelease *ShipRelease `json:"finalizeUploadedRelease"`
}

type GraphQLResponseUploadRelease struct {
	Data   ShipReleaseUploadData `json:"data,omitempty"`
	Errors []graphql.GQLError    `json:"errors,omitempty"`
}

type ShipReleaseUploadData struct {
	ShipPendingReleaseData *ShipPendingReleaseData `json:"uploadRelease"`
}

type ShipPendingReleaseData struct {
	UploadURI string `json:"uploadUri"`
	UploadID  string `json:"id"`
}

type GraphQLResponseLintRelease struct {
	Data   *ShipReleaseLintData `json:"data,omitempty"`
	Errors []graphql.GQLError   `json:"errors,omitempty"`
}

type ShipReleaseLintData struct {
	Messages []types.LintMessage `json:"lintRelease"`
}


const listReleasesQuery = `
query allReleases($appId: ID!) {
  allReleases(appId: $appId) {
    id
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
  }
}`

func (c *GraphQLClient) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	response := GraphQLResponseListReleases{}

	request := graphql.Request{
		Query: listReleasesQuery,
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
	for _, shipRelease := range response.Data.ShipReleases {
		activeChannels := make([]types.Channel, 0, 0)
		createdAt, err := util.ParseTime(shipRelease.CreatedAt)
		if err != nil {
			return nil, err
		}

		for _, shipReleaseChannel := range shipRelease.Channels {
			activeChannel := types.Channel{
				ID:   shipReleaseChannel.ID,
				Name: shipReleaseChannel.Name,
			}
			activeChannels = append(activeChannels, activeChannel)
		}

		releaseInfo := types.ReleaseInfo{
			AppID:          appID,
			CreatedAt:      createdAt.In(location),
			EditedAt:       time.Now(),
			Editable:       false,
			Sequence:       shipRelease.Sequence,
			Version:        "ba",
			ActiveChannels: activeChannels,
		}

		releaseInfos = append(releaseInfos, releaseInfo)
	}

	return releaseInfos, nil
}

const uploadReleaseQuery = `
mutation uploadRelease($appId: ID!) {
  uploadRelease(appId: $appId) {
    id
    uploadUri
  }
}`

const finalizeUploadedReleaseQuery = `
mutation finalizeUploadedRelease($appId: ID! $uploadId: String) {
  finalizeUploadedRelease(appId: $appId, uploadId: $uploadId) {
    id
    sequence
    spec
    created
    releaseNotes
  }
}`

func (c *GraphQLClient) CreateRelease(appID string, yaml string) (*types.ReleaseInfo, error) {
	response := GraphQLResponseUploadRelease{}

	request := graphql.Request{
		Query: uploadReleaseQuery,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "replicated-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml)); err != nil {
		return nil, err
	}
	tmpFile.Close()

	if err := util.UploadFile(tmpFile.Name(), response.Data.ShipPendingReleaseData.UploadURI); err != nil {
		return nil, err
	}

	request = graphql.Request{
		Query: finalizeUploadedReleaseQuery,
		Variables: map[string]interface{}{
			"appId":    appID,
			"uploadId": response.Data.ShipPendingReleaseData.UploadID,
		},
	}

	// call finalize release
	finalizeResponse := GraphQLResponseFinalizeRelease{}

	if err := c.ExecuteRequest(request, &finalizeResponse); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	if finalizeResponse.Data == nil || finalizeResponse.Data.ShipRelease == nil {
		return nil, fmt.Errorf("ship release not present in finalize response %+v", finalizeResponse)
	}

	createdAt, err := util.ParseTime(finalizeResponse.Data.ShipRelease.CreatedAt)
	if err != nil {
		return nil, err
	}

	releaseInfo := types.ReleaseInfo{
		AppID:     appID,
		CreatedAt: createdAt.In(location),
		EditedAt:  time.Now(),
		Editable:  false,
		Sequence:  finalizeResponse.Data.ShipRelease.Sequence,
		Version:   "ba",
	}

	return &releaseInfo, nil
}

const updateShipRelease = `
mutation updateRelease($appId: ID!, $spec: String!, $sequence: Int) {
    updateRelease(appId: $appId, spec: $spec, sequence: $sequence) {
      id
    }
  }`

func (c *GraphQLClient) UpdateRelease(appID string, sequence int64, yaml string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: updateShipRelease,
		Variables: map[string]interface{}{
			"appId":    appID,
			"sequence": sequence,
			"spec":     yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil

}

const promoteShipReleaseQuery = `
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $releaseNotes: String, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, releaseNotes: $releaseNotes, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`

func (c *GraphQLClient) PromoteRelease(appID string, sequence int64, label string, notes string, channelIDs ...string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: promoteShipReleaseQuery,
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

const lintReleaseQuery = `
mutation lintRelease($appId: ID!, $spec: String!) {
  lintRelease(appId: $appId, spec: $spec) {
    rule
    type
    positions {
      path
      start {
        position
        line
        column
      }
      end {
        position
        line
        column
      }
    }
  }
}`

func (c *GraphQLClient) LintRelease(appID string, yaml string) ([]types.LintMessage, error) {
	response := GraphQLResponseLintRelease{}

	request := graphql.Request{
		Query: lintReleaseQuery,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return response.Data.Messages, nil
}
