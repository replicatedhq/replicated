package kotsclient

import (
	"fmt"

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
	KotsReleaseData KotsReleaseSequence `json:"createKotsRelease"`
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

	// request = graphql.Request{
	// 	Query: `
	// 	mutation updateKotsRelease($appId: ID!, $spec: String!, $sequence: Int) {
	// 	  updateKotsRelease(appId: $appId, spec: $spec, sequence: $sequence) {
	// 		sequence
	// 	  }
	// 	}
	//   `,
	// 	Variables: map[string]interface{}{
	// 		"appId":    appID,
	// 		"spec":     multiyaml,
	// 		"sequence": response.Data.KotsReleaseData.Sequence,
	// 	},
	// }

	// finalizeKotsReleaseCreate := GraphQLResponseKotsUpdateRelease{}

	// if err := c.ExecuteRequest(request, &finalizeKotsReleaseCreate); err != nil {
	// 	return nil, err
	// }

	releaseInfo := types.ReleaseInfo{
		AppID:    appID,
		Sequence: response.Data.KotsReleaseData.Sequence,
	}

	return &releaseInfo, nil
}
