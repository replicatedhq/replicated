package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_AppCreate(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-app-create-token")
		client := VendorV3Client{HTTPClient: *api}

		createdApp, err := client.CreateKOTSApp("app-create-1")
		assert.Nil(t, err)

		assert.Equal(t, "app-create-1", createdApp.Name)
		assert.Equal(t, "replicated-cli-app-create", createdApp.TeamId)
		assert.Equal(t, true, createdApp.IsKotsApp)
		assert.Equal(t, false, createdApp.IsArchived)
		assert.Len(t, createdApp.Channels, 3)

		return nil
	}

	pact.AddInteraction().
		Given("Add a kots app").
		UponReceiving("A request to add a kots app").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/app"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-app-create-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name": dsl.String("app-create-1"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"app": map[string]interface{}{
					"channels":    dsl.EachLike(map[string]interface{}{}, 3),
					"created":     dsl.Timestamp(),
					"description": dsl.String(""),
					"id":          dsl.Term("2HKT3v84IjvCPSH03F3Hlg0Kpj6", ksuidRegex),
					"isArchived":  false,
					"isKotsApp":   true,
					"name":        dsl.String("app-create-1"),
					"renamedAt":   nil,
					"slug":        dsl.String("app-create-1"),
					"teamId":      dsl.String("replicated-cli-app-create"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
