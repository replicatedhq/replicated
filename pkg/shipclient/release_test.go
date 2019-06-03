package shipclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		request := GraphQLRequest{
			Query: uploadReleaseQuery,
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

		err = c.ExecuteRequest(request, &response)
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
				"query":         uploadReleaseQuery,
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

// TODO: once there is a fixture for providing s3 files add a pact that's meant to succeed
func Test_UploadRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		request := GraphQLRequest{
			Query: finalizeUploadedReleaseQuery,
			Variables: map[string]interface{}{
				"appId":    "ship-app-1",
				"uploadId": "upload-id-notexist",
			},
		}

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		response := GraphQLResponseFinalizeRelease{}

		err = c.executeRequest(request, &response)
		assert.Nil(t, err)

		// the upload does not exist, so we should expect an error
		assert.Len(t, response.Errors, 1)

		return nil
	}

	pact.AddInteraction().
		Given("finalize an uploaded release that does not exist for ship-app-1").
		UponReceiving("A request to finalize a nonexistent uploaded release for ship-app-1").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("basic-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         finalizeUploadedReleaseQuery,
				"variables": map[string]interface{}{
					"appId":    "ship-app-1",
					"uploadId": "upload-id-notexist",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"finalizeUploadedRelease": nil,
				},
				"errors": []map[string]interface{}{
					{
						"locations": []map[string]interface{}{
							{
								"line":   dsl.Like(dsl.Integer()),
								"column": dsl.Like(dsl.Integer()),
							},
						},
						"path": []string{
							"finalizeUploadedRelease",
						},
						"message": dsl.Like(dsl.String("Whoops, it looks like something is not working quite right. Please try back in a few minutes.")),
						"code":    dsl.Like(dsl.String("internal_server_error")),
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

		request := graphql.Request{
			Query: `
mutation promoteShipRelease($appId: ID!, $sequence: Int, $channelIds: [String], $versionLabel: String!, $troubleshootSpecId: ID!) {
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,

			Variables: map[string]interface{}{
				"appId":              "ship-app-1",
				"sequence":           1,
				"versionLabel":       "",
				"troubleshootSpecId": "",
				"channelIds":         []string{"ship-app-nightly"},
			},
		}

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		response := GraphQLResponseUploadRelease{}

		err = c.ExecuteRequest(request, &response)
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
  promoteShipRelease(appId: $appId, sequence: $sequence, channelIds: $channelIds, versionLabel: $versionLabel, troubleshootSpecId: $troubleshootSpecId) {
    id
  }
}`,
				"variables": map[string]interface{}{
					"appId":              "ship-app-1",
					"sequence":           1,
					"versionLabel":       "",
					"troubleshootSpecId": "",
					"channelIds":         []string{"ship-app-nightly"},
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

		err = c.PromoteRelease("ship-app-1", 1, "versionHere", "notesHere", "ship-app-beta")
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
				"query":         promoteShipReleaseQuery,
				"variables": map[string]interface{}{
					"appId":              "ship-app-1",
					"sequence":           1,
					"versionLabel":       "versionHere",
					"releaseNotes":       "notesHere",
					"troubleshootSpecId": "",
					"channelIds":         []string{"ship-app-beta"},
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

func Test_ListReleaseActual(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		releases, err := c.ListReleases("ship-app-1")
		assert.Nil(t, err)
		assert.Len(t, releases, 1)

		return nil
	}

	pact.AddInteraction().
		Given("list releases for ship-app-1").
		UponReceiving("A real request to list releases for ship-app-1").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("basic-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         listReleasesQuery,
				"variables": map[string]interface{}{
					"appId": "ship-app-1",
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"allReleases": []map[string]interface{}{
						{
							"id":       dsl.Like(dsl.String("generated uuid")),
							"sequence": dsl.Like(dsl.Integer()),
							"created":  dsl.Like(dsl.String("Tue Nov 10 2009 23:00:00 UTC")),
						},
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_LintReleaseActual(t *testing.T) {
	lintYaml := `
assets:
  v1:
    - inline:
        contents: |
          #!/bin/bash
          echo "installing nothing"
          echo "config option: {{repl ConfigOption "test_option" }}"
        dest: ./scripts/install.sh
        mode: 0777
intentional_breakage: {}
config:
  v1:
    - name: test_options
      title: Test Options
      description: testing testing 123
      items:
      - name: test_option
        title: Test Option
        default: abc123_test-option-value
        type: text
lifecycle:
  v1:
    - render: {}
`

	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/graphql", pact.Server.Port)

		uri, err := url.Parse(u)
		assert.Nil(t, err)
		c := &GraphQLClient{
			GQLServer: uri,
			Token:     "basic-read-write-token",
		}

		lintMessages, err := c.LintRelease("ship-app-1", lintYaml)
		assert.Nil(t, err)
		assert.Len(t, lintMessages, 1)

		return nil
	}

	pact.AddInteraction().
		Given("lint releases for ship-app-1").
		UponReceiving("A real request to lint a release for ship-app-1").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/graphql"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("basic-read-write-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"operationName": "",
				"query":         lintReleaseQuery,
				"variables": map[string]interface{}{
					"appId": "ship-app-1",
					"spec":  lintYaml,
				},
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"data": map[string]interface{}{
					"lintRelease": []map[string]interface{}{
						{
							"rule": dsl.Like(dsl.String("generated")),
							"type": dsl.Like(dsl.String("generated")),
							"positions": []map[string]interface{}{
								{
									"path": dsl.Like(dsl.String("generated")),
									"start": map[string]interface{}{
										"position": dsl.Like(dsl.Integer()),
										"line":     dsl.Like(dsl.Integer()),
										"column":   dsl.Like(dsl.Integer()),
									},
									"end": map[string]interface{}{
										"position": dsl.Like(dsl.Integer()),
										"line":     dsl.Like(dsl.Integer()),
										"column":   dsl.Like(dsl.Integer()),
									},
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
