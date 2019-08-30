package kotsclient

import (
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type KotsReleaseData struct {
	Sequence int64 `json:"sequence"`
}

var createKotsRelease = `
mutation createKotsRelease($appId: ID!, $spec: String!) {
  createKotsRelease(appId: $appId, spec: $spec) {
	sequence
  }
}
`

func (c *GraphQLClient) CreateRelease(appID string, yaml string) (*types.ReleaseInfo, error) {
	response := KotsReleaseData{}

	request := graphql.Request{
		Query: createKotsRelease,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	releaseInfo := types.ReleaseInfo{
		AppID:    appID,
		Sequence: response.Sequence,
	}

	return &releaseInfo, nil
}
