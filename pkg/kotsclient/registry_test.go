package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRegistryDockerHubPassword(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := VendorV3Client{HTTPClient: *api}

		req := AddKOTSRegistryRequest{
			Provider:       "dockerhub",
			Endpoint:       "index.docker.io",
			AuthType:       "password",
			Username:       "test",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "index.docker.io", response.Endpoint)
		assert.Equal(t, "dockerhub", response.Provider)
		assert.Equal(t, "password", response.AuthType)

		return nil
	}

	pact.AddInteraction().
		Given("Add a dockerhub external registry using a password").
		UponReceiving("A request to add a dockerhub external registry using authtype password").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("dockerhub"),
				"endpoint":       dsl.String("index.docker.io"),
				"authType":       dsl.String("password"),
				"username":       dsl.String("test"),
				"password":       dsl.String("test"),
				"skipValidation": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"external_registry": map[string]interface{}{
					"provider": dsl.String("dockerhub"),
					"endpoint": dsl.String("index.docker.io"),
					"authType": dsl.String("password"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateRegistryDockerHubAccessToken(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := VendorV3Client{HTTPClient: *api}

		req := AddKOTSRegistryRequest{
			Provider:       "dockerhub",
			Endpoint:       "index.docker.io",
			AuthType:       "token",
			Username:       "test",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "index.docker.io", response.Endpoint)
		assert.Equal(t, "dockerhub", response.Provider)
		assert.Equal(t, "token", response.AuthType)

		return nil
	}

	pact.AddInteraction().
		Given("Add a dockerhub external registry using a token").
		UponReceiving("A request to add a dockerhub external registry using authtype access token").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("dockerhub"),
				"endpoint":       dsl.String("index.docker.io"),
				"authType":       dsl.String("token"),
				"username":       dsl.String("test"),
				"password":       dsl.String("test"),
				"skipValidation": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"external_registry": map[string]interface{}{
					"provider": dsl.String("dockerhub"),
					"endpoint": dsl.String("index.docker.io"),
					"authType": dsl.String("token"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateRegistryECR(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := VendorV3Client{HTTPClient: *api}

		req := AddKOTSRegistryRequest{
			Provider:       "ecr",
			Endpoint:       "0000000000.dkr.ecr.us-east-2.amazonaws.com",
			AuthType:       "accesskey",
			Username:       "test",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "0000000000.dkr.ecr.us-east-2.amazonaws.com", response.Endpoint)
		assert.Equal(t, "ecr", response.Provider)
		assert.Equal(t, "accesskey", response.AuthType)

		return nil
	}

	pact.AddInteraction().
		Given("Add an ecr external registry using a accesskey").
		UponReceiving("A request to add an ecr external registry using auth type accesskey").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("ecr"),
				"endpoint":       dsl.String("0000000000.dkr.ecr.us-east-2.amazonaws.com"),
				"authType":       dsl.String("accesskey"),
				"username":       dsl.String("test"),
				"password":       dsl.String("test"),
				"skipValidation": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"external_registry": map[string]interface{}{
					"provider": dsl.String("ecr"),
					"endpoint": dsl.String("0000000000.dkr.ecr.us-east-2.amazonaws.com"),
					"authType": dsl.String("accesskey"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateRegistryGCR(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := VendorV3Client{HTTPClient: *api}

		req := AddKOTSRegistryRequest{
			Provider:       "gcr",
			Endpoint:       "gcr.io",
			AuthType:       "serviceaccount",
			Username:       "_json_key",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "gcr.io", response.Endpoint)
		assert.Equal(t, "gcr", response.Provider)
		assert.Equal(t, "serviceaccount", response.AuthType)

		return nil
	}

	pact.AddInteraction().
		Given("Add an gcr external registry using a serviceaccount").
		UponReceiving("A request to add an rcr external registry using auth type serviceaccount").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("gcr"),
				"endpoint":       dsl.String("gcr.io"),
				"authType":       dsl.String("serviceaccount"),
				"username":       dsl.String("_json_key"),
				"password":       dsl.String("test"),
				"skipValidation": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"external_registry": map[string]interface{}{
					"provider": dsl.String("gcr"),
					"endpoint": dsl.String("gcr.io"),
					"authType": dsl.String("serviceaccount"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_RemoveRegistryDockerHubPassword(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-rm-registry-token")
		client := VendorV3Client{HTTPClient: *api}

		err = client.RemoveKOTSRegistry("index.docker.io")
		assert.Nil(t, err)

		registries, err := client.ListRegistries()
		assert.Nil(t, err)

		assert.Equal(t, 0, len(registries))

		return nil
	}

	pact.AddInteraction().
		Given("Remove a dockerhub external registry using a password").
		UponReceiving("A request to remove a dockerhub external registry using authtype password").
		WithRequest(dsl.Request{
			Method: "DELETE",
			Path:   dsl.String("/v3/external_registry/index.docker.io"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 204,
		})

	pact.AddInteraction().
		Given("List registries after deleting the only one").
		UponReceiving("A request to list registries after deleting them").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/external_registries"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-rm-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"external_registries": []interface{}{},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
