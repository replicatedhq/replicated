package kotsclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"

	"github.com/replicatedhq/replicated/pkg/graphql"
)

func Test_ListKotsCustomersActual(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		d := &graphql.Client{
			GQLServer: uri,
			Token:     "list-kots-customers-write-token",
		}

		c := &GraphQLClient{GraphQLClient: d}

		releases, err := c.ListCustomers("list-kots-customers")
		assert.Nil(t, err)
		assert.Len(t, releases, 2)

		return nil
	}

	pact.AddInteraction().
		Given("list customers for list-kots-customers").
		UponReceiving("A real request to list customers for list-kots-customers").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("list-kots-customers-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         kotsListCustomers,
				"variables": map[string]interface{}{
					"appId":   "list-kots-customers",
					"appType": "kots",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"customers": map[string]interface{}{
						"customers": []map[string]interface{}{
							{
								"id": "im-fake",
							},
							{
								"id": "im-also-fake",
							},
						},
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
