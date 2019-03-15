package shipclient

import (
	"time"

	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListReleases struct {
	Data   *ShipReleasesData `json:"data,omitempty"`
	Errors []GraphQLError    `json:"errors,omitempty"`
}

type ShipReleasesData struct {
	ShipReleases []*ShipRelease `json:"allReleases"`
}

type ShipRelease struct {
	ID           string `json:"id"`
	Sequence     int64  `json:"sequence"`
	CreatedAt    string `json:"created"`
	ReleaseNotes string `json:"releaseNotes"`
}

type GraphQLResponseCreateRelease struct {
	Data   *ShipReleaseCreateData `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type ShipReleaseCreateData struct {
	ShipRelease *ShipRelease `json:"createRelease"`
}

type GraphQLResponseLintRelease struct {
	Data   *ShipReleaseLintData `json:"data,omitempty"`
	Errors []GraphQLError       `json:"errors,omitempty"`
}

type ShipReleaseLintData struct {
	Messages []*ShipLintMessage `json:"lintRelease"`
}

type ShipLintMessage struct {
	Rule      string              `json:"rule"`
	Type      string              `json:"type"`
	Positions []*ShipLintPosition `json:"positions"`
}

type ShipLintPosition struct {
	Path  string                `json:"path"`
	Start *ShipLintLinePosition `json:"start"`
	End   *ShipLintLinePosition `json:"end"`
}

type ShipLintLinePosition struct {
	Position int64 `json:"position"`
	Line     int64 `json:"line"`
	Column   int64 `json:"column"`
}

func (c *GraphQLClient) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	response := GraphQLResponseListReleases{}

	request := GraphQLRequest{
		Query: `
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
}`,
		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	releaseInfos := make([]types.ReleaseInfo, 0, 0)
	for _, shipRelease := range response.Data.ShipReleases {
		createdAt, err := time.Parse("Mon Jan 02 2006 15:04:05 MST-0700 (MST)", shipRelease.CreatedAt)
		if err != nil {
			return nil, err
		}
		releaseInfo := types.ReleaseInfo{
			AppID:     appID,
			CreatedAt: createdAt.In(location),
			EditedAt:  time.Now(),
			Editable:  false,
			Sequence:  shipRelease.Sequence,
			Version:   "ba",
		}

		releaseInfos = append(releaseInfos, releaseInfo)
	}

	return releaseInfos, nil
}

func (c *GraphQLClient) CreateRelease(appID string, yaml string) (*types.ReleaseInfo, error) {
	response := GraphQLResponseCreateRelease{}

	request := GraphQLRequest{
		Query: `
mutation createRelease($appId: ID!, $spec: String!) {
  createRelease(appId: $appId, spec: $spec) {
    id
    sequence
    spec
    created
    releaseNotes
  }
}`,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse("Mon Jan 02 2006 15:04:05 MST-0700 (MST)", response.Data.ShipRelease.CreatedAt)
	if err != nil {
		return nil, err
	}
	releaseInfo := types.ReleaseInfo{
		AppID:     appID,
		CreatedAt: createdAt.In(location),
		EditedAt:  time.Now(),
		Editable:  false,
		Sequence:  response.Data.ShipRelease.Sequence,
		Version:   "ba",
	}

	return &releaseInfo, nil
}

func (c *GraphQLClient) UpdateRelease(appID string, sequence int64, yaml string) error {
	return nil
}

func (c *GraphQLClient) PromoteRelease(appID string, sequence int64, label string, notes string, channelIDs ...string) error {
	response := GraphQLResponseErrorOnly{}

	request := GraphQLRequest{
		Query: `
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $releaseNotes: String, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, releaseNotes: $releaseNotes, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,
		Variables: map[string]interface{}{
			"appId":              appID,
			"sequence":           sequence,
			"versionLabel":       label,
			"releaseNotes":       notes,
			"troubleshootSpecId": "",
			"channelIds":         channelIDs,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return err
	}

	return nil
}

func (c *GraphQLClient) LintRelease(appID string, yaml string) ([]types.LintMessage, error) {
	response := GraphQLResponseLintRelease{}

	request := GraphQLRequest{
		Query: `
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
}`,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	lintMessages := make([]types.LintMessage, 0, 0)
	for _, message := range response.Data.Messages {
		positions := make([]*types.LintPosition, 0, 0)
		for _, lintPosition := range message.Positions {
			position := types.LintPosition{
				Path: lintPosition.Path,
				Start: &types.LintLinePosition{
					Position: lintPosition.Start.Position,
					Column:   lintPosition.Start.Column,
					Line:     lintPosition.Start.Line,
				},
				End: &types.LintLinePosition{
					Position: lintPosition.End.Position,
					Column:   lintPosition.End.Column,
					Line:     lintPosition.End.Line,
				},
			}

			positions = append(positions, &position)
		}

		lintMessage := types.LintMessage{
			Rule:      message.Rule,
			Type:      message.Type,
			Positions: positions,
		}

		lintMessages = append(lintMessages, lintMessage)
	}

	return lintMessages, nil
}
