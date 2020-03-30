package kotsclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"

	"github.com/replicatedhq/replicated/pkg/graphql"
)

func Test_ListKotsReleasesActual(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)

		d := &graphql.Client{
			GQLServer: uri,
			Token:     "all-kots-releases-read-write-token",
		}

		c := &GraphQLClient{
			GraphQLClient:    d,
			KurlDotSHAddress: "",
		}

		releases, err := c.ListReleases("all-kots-releases")
		assert.Nil(t, err)
		assert.Len(t, releases, 2)

		return nil
	}

	pact.AddInteraction().
		Given("list releases for all-kots-releases").
		UponReceiving("A real request to list releases for all-kots-releases").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("all-kots-releases-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         allKotsReleases,
				"variables": map[string]interface{}{
					"appId": "all-kots-releases",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"allKotsReleases": []map[string]interface{}{
						{
							"sequence": 2,
							"created":  dsl.Like(dsl.String("Tue Nov 10 2009 23:00:00 UTC")),
							"channels": []map[string]interface{}{
								{
									"id":             "all-kots-releases-beta",
									"name":           "Beta",
									"currentVersion": "1.0.1",
									"numReleases":    1,
								},
								{
									"id":             "all-kots-releases-nightly",
									"name":           "Nightly",
									"currentVersion": "1.0.1",
									"numReleases":    2,
								},
							},
						},
						{
							"sequence": 1,
							"created":  dsl.Like(dsl.String("Tue Nov 10 2009 23:00:00 UTC")),
							"channels": []map[string]interface{}{
								{
									"id":             "all-kots-releases-test",
									"name":           "Test",
									"currentVersion": "1.0.0",
									"numReleases":    1,
								},
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
