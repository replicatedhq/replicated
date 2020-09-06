package kotsclient

import (
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

type GraphQLResponseKotsGetRelease struct {
	Data   KOTSReleaseResponseData `json:"data,omitempty"`
	Errors []graphql.GQLError      `json:"errors,omitempty"`
}

type KOTSReleaseResponseData struct {
	KotsReleaseData *KOTSReleaseWithSpec `json:"kotsReleaseForSequence"`
}

type KOTSReleaseWithSpec struct {
	Sequence     int64  `json:"sequence"`
	ReleaseNotes string `json:"releaseNotes"`
	Spec         string `json:"spec"`
}

const kotsQueryGetReleaseForSequence = `
query kotsReleaseForSequence($appId: ID!, $sequence: Int!) {
  kotsReleaseForSequence(appId: $appId, sequence: $sequence) {
    sequence
    releaseNotes
    spec
  }
}`

func (c *GraphQLClient) GetRelease(appID string, sequence int64) (*releases.AppRelease, error) {
	response := GraphQLResponseKotsGetRelease{}

	request := graphql.Request{
		Query: kotsQueryGetReleaseForSequence,
		Variables: map[string]interface{}{
			"appId": appID,
			"sequence": sequence,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}
	return &releases.AppRelease{
		Config:    response.Data.KotsReleaseData.Spec,
		Sequence:  response.Data.KotsReleaseData.Sequence,
	}, nil
}
