package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateInstaller(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "repl-cli-create-installer-token")
		client := VendorV3Client{HTTPClient: *api}

		installerSpec, err := client.CreateInstaller("repl-cli-create-installer-app", testInstallerYAML)
		assert.NoError(t, err)

		assert.Equal(t, "repl-cli-create-installer-app", installerSpec.AppID)
		return nil
	}

	pact.AddInteraction().
		Given("Create KOTS installer").
		UponReceiving("A request to create a kots installer").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/app/repl-cli-create-installer-app/installer"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("repl-cli-create-installer-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"yaml": testInstallerYAML,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"installer": map[string]interface{}{
					"appId": "repl-cli-create-installer-app",
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
