package shipclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		request := GraphQLRequest{
			Query: `
mutation uploadRelease($appId: ID!) {
  uploadRelease(appId: $appId) {
    id
    uploadUri
  }
}`,
			Variables: map[string]interface{}{
				"appId": "ship-app-1",
			},
		}

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		response := GraphQLResponseUploadRelease{}

		err = c.executeRequest(request, &response)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Prepare to upload a release for ship-app-1").
		UponReceiving("A request to upload a new release for ship-app-1").
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
mutation uploadRelease($appId: ID!) {
  uploadRelease(appId: $appId) {
    id
    uploadUri
  }
}`,
				"variables": map[string]interface{}{
					"appId": "ship-app-1",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"uploadRelease": map[string]interface{}{
						"id":        dsl.Like(dsl.String("generated")),
						"uploadUri": dsl.Like(dsl.String("generated")),
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_PromoteReleaseMinimal(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		request := GraphQLRequest{
			Query: `
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,

			Variables: map[string]interface{}{
				"appId":              "ship-app-1",
				"sequence":           1,
				"troubleshootSpecId": "",
				"channelIds":         []string{"Nightly"},
			},
		}

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		response := GraphQLResponseUploadRelease{}

		err = c.executeRequest(request, &response)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Prepare to (mock) promote a release for ship-app-1").
		UponReceiving("A mocked minimal request to promote a new release for ship-app-1").
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
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,
				"variables": map[string]interface{}{
					"appId":              "ship-app-1",
					"sequence":           1,
					"troubleshootSpecId": "",
					"channelIds":         []string{"Nightly"},
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"promoteShipRelease": map[string]interface{}{
						"id": dsl.Like(dsl.String("generated uuid")),
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_PromoteReleaseActual(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		err = c.PromoteRelease("ship-app-1", 1, "versionHere", "notesHere", "Nightly")
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Prepare to promote a release for ship-app-1").
		UponReceiving("A real request to promote a new release for ship-app-1").
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
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $releaseNotes: String, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, releaseNotes: $releaseNotes, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,
				"variables": map[string]interface{}{
					"appId":              "ship-app-1",
					"sequence":           1,
					"versionLabel":       "versionHere",
					"releaseNotes":       "notesHere",
					"troubleshootSpecId": "",
					"channelIds":         []string{"Nightly"},
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"promoteShipRelease": map[string]interface{}{
						"id": dsl.Like(dsl.String("generated uuid")),
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
