package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_ListApps(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-list-apps-token")
		client := VendorV3Client{HTTPClient: *api}

		apps, err := client.ListApps()
		assert.Nil(t, err)

		assert.Len(t, apps, 1)

		app := apps[0]

		assert.Equal(t, "replicated-cli-list-apps-app", app.App.ID)
		assert.Equal(t, "replicated-cli-list-apps-app", app.App.Slug)
		assert.Equal(t, "Replicated CLI List Apps App", app.App.Name)
		assert.Equal(t, "kots", app.App.Scheduler)

		assert.Len(t, app.Channels, 1)

		return nil
	}

	pact.AddInteraction().
		Given("List KOTS apps").
		UponReceiving("A request to list kots apps").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/apps"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-apps-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"apps": []map[string]interface{}{
					{
						"name":     "Replicated CLI List Apps App",
						"slug":     "replicated-cli-list-apps-app",
						"id":       "replicated-cli-list-apps-app",
						"channels": dsl.EachLike(map[string]interface{}{}, 1),
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_RemoveApp(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-rm-app-token")
		client := VendorV3Client{HTTPClient: *api}

		err = client.DeleteKOTSApp("replicated-cli-rm-app-app")
		assert.Nil(t, err)

		apps, err := client.ListApps()
		assert.Nil(t, err)

		assert.Len(t, apps, 0)

		return nil
	}

	pact.AddInteraction().
		Given("Delete KOTS app").
		UponReceiving("A request to delete a kots app").
		WithRequest(dsl.Request{
			Method: "DELETE",
			Path:   dsl.String("/v3/app/replicated-cli-rm-app-app"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-app-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body:   map[string]interface{}{},
		})

	pact.AddInteraction().
		Given("List KOTS apps after deleting").
		UponReceiving("A request to list kots apps after deleting").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/apps"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-app-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body:   map[string]interface{}{},
		})
	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
