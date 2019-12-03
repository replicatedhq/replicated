package kotsclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func Test_ListKotsReleasesActual(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		d := &graphql.Client{
			GQLServer: uri,
			Token:     "kots-app-read-write-token",
		}

		c := &GraphQLClient{GraphQLClient: d}

		releases, err := c.ListReleases("kots-app")
		assert.Nil(t, err)
		assert.Len(t, releases, 1)

		return nil
	}

	pact.AddInteraction().
		Given("list releases for kots-app").
		UponReceiving("A real request to list releases for kots-app").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("kots-app-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         allKotsReleases,
				"variables": map[string]interface{}{
					"appId": "kots-app",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"allKotsReleases": dsl.EachLike(map[string]interface{}{
						"sequence": dsl.Like(dsl.Integer()),
						"created":  dsl.Like(dsl.String("Tue Nov 10 2009 23:00:00 UTC")),
						"channels": dsl.EachLike(map[string]interface{}{
							"id":             dsl.Like(dsl.String("id")),
							"name":           dsl.Like(dsl.String("channel name")),
							"currentVersion": dsl.Like(dsl.String("1.2.3")),
						}, 0),
					}, 1),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
