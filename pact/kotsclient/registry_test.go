package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	realkotsclient "github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRegistryDockerHubPassword(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "dockerhub",
			Endpoint:       "index.docker.io", // explicitly did not set slug to assert api fall back to using the endpoint as the slug
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
		assert.Equal(t, "index.docker.io", response.Slug)

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
				"slug":           dsl.String(""),
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
					"slug":     dsl.String("index.docker.io"),
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "dockerhub",
			Endpoint:       "index.docker.io",
			Slug:           "token-test",
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
		assert.Equal(t, "token-test", response.Slug)

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
				"slug":           dsl.String("token-test"),
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
					"slug":     dsl.String("token-test"),
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "ecr",
			Endpoint:       "0000000000.dkr.ecr.us-east-2.amazonaws.com",
			Slug:           "ecr-test-registry",
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
		assert.Equal(t, "ecr-test-registry", response.Slug)

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
				"slug":           dsl.String("ecr-test-registry"),
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
					"slug":     dsl.String("ecr-test-registry"),
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "gcr",
			Endpoint:       "gcr.io",
			Slug:           "gcr-test-registry",
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
		assert.Equal(t, "gcr-test-registry", response.Slug)

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
				"slug":           dsl.String("gcr-test-registry"),
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
					"slug":     dsl.String("gcr-test-registry"),
					"authType": dsl.String("serviceaccount"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateRegistryGARServiceAccount(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "gar",
			Endpoint:       "pkg.dev",
			Slug:           "gar-serviceaccount-test",
			AuthType:       "serviceaccount",
			Username:       "_json_key",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "pkg.dev", response.Endpoint)
		assert.Equal(t, "gar", response.Provider)
		assert.Equal(t, "serviceaccount", response.AuthType)
		assert.Equal(t, "gar-serviceaccount-test", response.Slug)

		return nil
	}

	pact.AddInteraction().
		Given("Add an gar external registry using a serviceaccount").
		UponReceiving("A request to add a gar external registry using auth type serviceaccount").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("gar"),
				"endpoint":       dsl.String("pkg.dev"),
				"slug":           dsl.String("gar-serviceaccount-test"),
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
					"provider": dsl.String("gar"),
					"endpoint": dsl.String("pkg.dev"),
					"slug":     dsl.String("gar-serviceaccount-test"),
					"authType": dsl.String("serviceaccount"),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_CreateRegistryGARAccessToken(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-add-registry-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		req := realkotsclient.AddKOTSRegistryRequest{
			Provider:       "gar",
			Endpoint:       "pkg.dev",
			Slug:           "gar-accesstoken-test",
			AuthType:       "token",
			Username:       "oauth2accesstoken",
			Password:       "test",
			SkipValidation: true,
		}
		response, err := client.AddKOTSRegistry(req)
		assert.Nil(t, err)

		assert.Equal(t, "pkg.dev", response.Endpoint)
		assert.Equal(t, "gar", response.Provider)
		assert.Equal(t, "token", response.AuthType)
		assert.Equal(t, "gar-accesstoken-test", response.Slug)

		return nil
	}

	pact.AddInteraction().
		Given("Add an gar external registry using an access token").
		UponReceiving("A request to add a gar external registry using auth type access token").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/external_registry"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-add-registry-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"provider":       dsl.String("gar"),
				"endpoint":       dsl.String("pkg.dev"),
				"slug":           dsl.String("gar-accesstoken-test"),
				"authType":       dsl.String("token"),
				"username":       dsl.String("oauth2accesstoken"),
				"password":       dsl.String("test"),
				"skipValidation": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"external_registry": map[string]interface{}{
					"provider": dsl.String("gar"),
					"endpoint": dsl.String("pkg.dev"),
					"slug":     dsl.String("gar-accesstoken-test"),
					"authType": dsl.String("token"),
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		err = client.RemoveKOTSRegistry("dockerhub-fixture")
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
			Path:   dsl.String("/v3/external_registry/dockerhub-fixture"),
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
				"external_registries": nil,
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
