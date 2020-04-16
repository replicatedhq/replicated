package kotsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/version"
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

func (c *GraphQLClient) LintRelease(appID, allKotsYamlsAsJson string) ([]types.LintMessage, error) {
	response := GraphQLResponseLintRelease{}

	request := graphql.Request{
		Query:         lintKotsRelease,
		OperationName: "lintKotsSpec",
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  allKotsYamlsAsJson,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, errors.Wrap(err, "execute request")
	}

	return response.Data.Messages, nil
}

// this is part of the gql client with plans to rename gql client to kotsclient
// and have endpoints for multiple release services included
func (c *GraphQLClient) LintReleaseBeta(data []byte) ([]types.LintMessage, error) {
	endpoint := "https://lint.replicated.com/v1/lint"

	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", endpoint, reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("User-Agent", fmt.Sprintf("Replicated/%s", version.Version()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Wrap(err, "non OK response from linter")
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response")
	}

	type betaResponse struct {
		LintExpressions []types.LintMessage `json:"lintExpressions"`
	}
	br := betaResponse{}

	if err := json.Unmarshal(msg, &br); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return br.LintExpressions, nil
}
