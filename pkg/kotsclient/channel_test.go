package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateChannel(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-create-channel-token")
		client := VendorV3Client{HTTPClient: *api}

		channel, err := client.CreateChannel("replicated-cli-create-channel-app", "New Channel", "Description")
		assert.Nil(t, err)

		// unstable exists as channel 0
		assert.Equal(t, "New Channel", channel.Name)
		assert.Equal(t, "Description", channel.Description)

		return nil
	}

	pact.AddInteraction().
		Given("Create KOTS app channel").
		UponReceiving("A request to create a kots app channel").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/app/replicated-cli-create-channel-app/channel"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-create-channel-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":        "New Channel",
				"description": "Description",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"channel": map[string]interface{}{
					"id":          dsl.Like("replicated-cli-create-channel-channel"),
					"name":        "New Channel",
					"description": "Description",
					"slug":        dsl.Like("new-channel"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_GetChannel(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-get-channel-token")
		client := VendorV3Client{HTTPClient: *api}

		channel, releases, err := client.GetChannel("replicated-cli-get-channel-app", "unstable")
		assert.Nil(t, err)

		assert.Equal(t, "Unstable", channel.Name)
		assert.Len(t, releases, 0) // we don't return the releases

		return nil
	}

	pact.AddInteraction().
		Given("Get KOTS app channel").
		UponReceiving("A request to get kots app channel").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-get-channel-app/channel/unstable"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-get-channel-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"channel": map[string]interface{}{
					"id":   dsl.Like("replicated-cli-get-channel-unstable"),
					"name": "Unstable",
					"releases": []map[string]interface{}{
						{
							"sequence":         1,
							"version":          "0.0.1",
							"channelSeqeuence": 1,
						},
						{
							"sequence":         2,
							"version":          "0.0.2",
							"channelSeqeuence": 2,
						},
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_ListChannels(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-list-channels-token")
		client := VendorV3Client{HTTPClient: *api}

		channels, err := client.ListChannels("replicated-cli-list-channels-app", "", "")
		assert.Nil(t, err)

		assert.Len(t, channels, 1)
		assert.Equal(t, "Unstable", channels[0].Name)

		return nil
	}

	pact.AddInteraction().
		Given("List KOTS app channels").
		UponReceiving("A request to list kots app channels").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-list-channels-app/channels"),
			Query: dsl.MapMatcher{
				"excludeDetail": dsl.String("true"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-channels-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"channels": []map[string]interface{}{
					{
						"id":   dsl.Like("replicated-cli-list-channels-unstable"),
						"name": "Unstable",
					},
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
