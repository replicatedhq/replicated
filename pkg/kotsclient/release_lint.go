package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseLintRelease struct {
	Data   *KOTSReleaseLintData `json:"data,omitempty"`
	Errors []graphql.GQLError   `json:"errors,omitempty"`
}

type KOTSReleaseLintData struct {
	Messages []types.LintMessage `json:"lintKotsSpec"`
}

const lintKotsRelease = `
query lintKotsSpec($appId: ID!, $spec: String!) {
  lintKotsSpec(appId: $appId, spec: $spec) {
    rule
    type
    message
    path
    positions {
      start {
        line
      }
    }
  }
}
`

func (c *HybridClient) LintRelease(appID, allKotsYamlsAsJson string) ([]types.LintMessage, error) {

	response := GraphQLResponseLintRelease{}

	request := graphql.Request{
		Query:         lintKotsRelease,
		OperationName: "lintKotsSpec",
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  allKotsYamlsAsJson,
		},
	}

	if err := c.ExecuteGraphQLRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute request")
	}

	return response.Data.Messages, nil

}
