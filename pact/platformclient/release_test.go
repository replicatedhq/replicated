package platformclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"

	swagger "github.com/replicatedhq/replicated/gen/go/v1"
	realplatformclient "github.com/replicatedhq/replicated/pkg/platformclient"
)

func Test_CreateRelease(t *testing.T) {
	var test = func() (err error) {
		appId := "cli-create-release-app-id"
		token := "cli-create-release-auth"

		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		client := realplatformclient.NewHTTPClient(u, token)

		release, err := client.CreateRelease(appId, "")
		assert.Nil(t, err)
		assert.Equal(t, true, release.Editable)
		return nil
	}

	pact.AddInteraction().
		Given("Create a release for cli-create-release-app-id").
		UponReceiving("A request to create a new release for cli-create-release-app-id").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v1/app/cli-create-release-app-id/release"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-release-auth"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"source":     "latest",
				"sourcedata": 0,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"Sequence":  dsl.Like(10),
				"Config":    dsl.Like(""),
				"Editable":  true,
				"CreatedAt": dsl.Like("2006-01-02T15:04:05Z"),
				"EditedAt":  dsl.Like("2006-01-02T15:04:05Z"),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateReleaseEmpty(t *testing.T) {
	var test = func() (err error) {
		appId := "cli-create-release-app-id"
		token := "cli-create-release-auth"
		u := fmt.Sprintf("http://localhost:%d/v1/app/%s/release", pact.Server.Port, appId)

		req, err := http.NewRequest("POST", u, bytes.NewReader(nil))
		assert.Nil(t, err)

		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)

		release := &swagger.AppReleaseInfo{}
		err = json.NewDecoder(resp.Body).Decode(release)
		assert.Nil(t, err)
		assert.Equal(t, true, release.Editable)
		return nil
	}

	pact.AddInteraction().
		Given("Empty create a release for cli-create-release-app-id").
		UponReceiving("An empty request to create a new release for cli-create-release-app-id").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v1/app/cli-create-release-app-id/release"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-release-auth"),
			},
			Body: nil,
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"Sequence":  dsl.Like(10),
				"Config":    dsl.Like(""),
				"Editable":  true,
				"CreatedAt": dsl.Like("2006-01-02T15:04:05Z"),
				"EditedAt":  dsl.Like("2006-01-02T15:04:05Z"),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetRelease(t *testing.T) {
	var test = func() (err error) {
		appId := "cli-create-release-app-id"
		token := "cli-create-release-auth"

		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)
		client := realplatformclient.NewHTTPClient(u, token)

		release, err := client.GetRelease(appId, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), release.Sequence)
		return nil
	}

	pact.AddInteraction().
		Given("Get a release for cli-create-release-app-id").
		UponReceiving("A request to get an existing release for cli-create-release-app-id").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v1/app/cli-create-release-app-id/2/properties"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("cli-create-release-auth"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"Sequence":  2, // mandated  by requesting sequence #2
				"Config":    dsl.Like("there might be a config here"),
				"Editable":  true,
				"CreatedAt": dsl.Like("2006-01-02T15:04:05Z"),
				"EditedAt":  dsl.Like("2006-01-02T15:04:05Z"),
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
