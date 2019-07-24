package shipclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func Test_CreateEntitlementSpec(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		request := graphql.Request{
			Query: `
mutation createEntitlementSpec($spec: String!, $name: String!, $appId: String!) {
  createEntitlementSpec(spec: $spec, name: $name, labels:[{key:"replicated.com/app", value:$appId}]) {
    id
    spec
    name
    createdAt
  }
}`,
			Variables: map[string]interface{}{
				"appId": "ship-app-1",
				"name":  "0.1.0",
				"spec":  "---\n- name: My Field\n  key: num_seats\n  description: Number of Seats\n  type: string\n  default: \"10\"\n  labels:\n    - owner=somePerson\n",
			},
		}

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		d := &graphql.Client{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		c := &GraphQLClient{GraphQLClient: d}

		response := GraphQLResponseUploadRelease{}

		err = c.ExecuteRequest(request, &response)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("A new entitlement spec to set for ship-team").
		UponReceiving("A request to create a new entitlement spec for ship team").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("basic-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query": `
mutation createEntitlementSpec($spec: String!, $name: String!, $appId: String!) {
  createEntitlementSpec(spec: $spec, name: $name, labels:[{key:"replicated.com/app", value:$appId}]) {
    id
    spec
    name
    createdAt
  }
}`,
				"variables": map[string]interface{}{
					"appId": "ship-app-1",
					"name":  "0.1.0",
					"spec":  "---\n- name: My Field\n  key: num_seats\n  description: Number of Seats\n  type: string\n  default: \"10\"\n  labels:\n    - owner=somePerson\n",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"createEntitlementSpec": map[string]interface{}{
						"id":        dsl.Like(dsl.String("generated")),
						"spec":      dsl.Like(dsl.String("generated")),
						"name":      dsl.String("0.1.0"),
						"createdAt": dsl.Like(dsl.String("2019-01-01T01:23:45.678Z")),
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
