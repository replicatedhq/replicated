package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_ClusterList(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		clientFunc func() error
	}{
		{
			name:  "List clusters using service account",
			token: "replicated-cli-tokens-sa-token",
			clientFunc: func() (err error) {
				u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

				api := platformclient.NewHTTPClient(u, "replicated-cli-tokens-sa-token")
				client := VendorV3Client{HTTPClient: *api}

				_, err = client.ListClusters(true, nil, nil)
				assert.Nil(t, err)

				return nil
			},
		},
		{
			name:  "List clusters using personal token",
			token: "replicated-cli-tokens-personal-token",
			clientFunc: func() (err error) {
				u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

				api := platformclient.NewHTTPClient(u, "replicated-cli-tokens-personal-token")
				client := VendorV3Client{HTTPClient: *api}

				_, err = client.ListClusters(true, nil, nil)
				assert.Nil(t, err)

				return nil
			},
		},
	}

	for _, test := range tests {
		pact.AddInteraction().
			Given(test.name).
			UponReceiving("A request to list clusters").
			WithRequest(dsl.Request{
				Method: "GET",
				Path:   dsl.String("/v3/clusters"),
				Headers: dsl.MapMatcher{
					"Authorization": dsl.String(test.token),
					"Content-Type":  dsl.String("application/json"),
				},
			}).
			WillRespondWith(dsl.Response{
				Status: 200,
				Body: map[string]interface{}{
					"totalClusters": dsl.Like(0),
				},
			})

		if err := pact.Verify(test.clientFunc); err != nil {
			t.Fatalf("Error on Verify test %s: %v", test.name, err)
		}
	}
}
