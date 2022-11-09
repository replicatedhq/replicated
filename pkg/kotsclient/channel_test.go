package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
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
					"id":          dsl.Term("2HKXWE5CM7bqkR5T2sKfVniMJfD", ksuidRegex),
					"name":        "New Channel",
					"description": "Description",
					"channelSlug": dsl.String("new-channel"),
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

func Test_RemoveChannels(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-rm-channel-token")
		client := VendorV3Client{HTTPClient: *api}

		err = client.ArchiveChannel("replicated-cli-rm-channel-app", "replicated-cli-rm-channel-beta")
		assert.Nil(t, err)

		channels, err := client.ListChannels("replicated-cli-rm-channel-app", "", "")
		assert.Nil(t, err)

		assert.Len(t, channels, 1)

		return nil
	}

	pact.AddInteraction().
		Given("Remove KOTS app channel").
		UponReceiving("A request to remove kots app channel").
		WithRequest(dsl.Request{
			Method: "DELETE",
			Path:   dsl.String("/v3/app/replicated-cli-rm-channel-app/channel/replicated-cli-rm-channel-beta"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-channel-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body:   map[string]interface{}{},
		})

	pact.AddInteraction().
		Given("List KOTS app channels").
		UponReceiving("A request to list kots app channels after removing one").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-rm-channel-app/channels"),
			Query: dsl.MapMatcher{
				"excludeDetail": dsl.String("true"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-channel-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"channels": []map[string]interface{}{
					{
						"id":   dsl.Like("replicated-cli-rm-channel-unstable"),
						"name": "Unstable",
					},
				},
			},
		})
	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_AddRemoveSemver(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-semver-channel-token")
		client := VendorV3Client{HTTPClient: *api}

		channel := channels.AppChannel{
			Name: "Unstable",
			Id:   "replicated-cli-semver-channel-unstable",
		}

		err = client.UpdateSemanticVersioning("replicated-cli-semver-channel-app", &channel, true)
		assert.Nil(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Add Semver to a KOTS app channel").
		UponReceiving("A request to add semver to kots app channel").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/app/replicated-cli-semver-channel-app/channel/replicated-cli-semver-channel-unstable"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-semver-channel-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"name":           "Unstable",
				"semverRequired": true,
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
