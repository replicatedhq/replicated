package kotsclient

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateGetRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-kots-release-token")
		client := VendorV3Client{HTTPClient: *api}

		releaseInfo, err := client.CreateRelease("replicated-cli-kots-release-app", testYAML)
		assert.NoError(t, err)

		assert.Equal(t, "replicated-cli-kots-release-app", releaseInfo.AppID)
		assert.Equal(t, int64(1), releaseInfo.Sequence)

		_, err = client.GetRelease("replicated-cli-kots-release-app", 1)
		assert.NoError(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Create KOTS release").
		UponReceiving("A request to create a kots release").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/app/replicated-cli-kots-release-app/release"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-kots-release-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"spec_gzip": "H4sIAAAAAAAA/6SSMW/bMBCFd/6KgzJkogRkK4F2SQp0aod0KwqBoS4RYeqO5Z2sGoH/e0HJjj1kKbLpRPG9p++dtdYUzCkGrzj0Psd+j0Uik4O79lN7Z8hP6KD5iaKNMTfmBkbVLK7rlmVpL3fbwFM3cJAu+7DzL5FerCfr8/ZBZOqunm0unLFoRDE35jI4A+Bz7ueS3Grkuu71tbrAPdNzfPmR63VoRhat0Ro4Hg1AYBJO2GvU9J9xd0+dzDlz0Rr5wHOxYRblCYt0kUR9SnZHvJA9oamRq39f8M8cC05Iuia/IvlGsfnyeQV5SvOIWn0EJBRE+gjOsAKxZyGzzc5Y2Do7IzIAJyzfLm8GlFDiCtPBV5K5IOgYBQaefKRVAqJA4Vn9U0JggsoGCHXhsmsNQFSc1v9+x/E9T4C9TzM6uL1UWls7MYFGk7RvxR6Pt5vOIaMDxb+6jSja58LBreMQJSd/6LcA9yOGHTx8f1zPAk+Tp8FBQeG0x75qbzU84HMkhJEX0BHhCmtdJfWRsAgsMSV4QggFayfgaQBRXxSHjxU3Zaa6NNbTYC+Ga4nnMwe/fpt/AQAA//9wVga6oAMAAA==",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 201,
			Body: map[string]interface{}{
				"release": map[string]interface{}{
					"appId":    "replicated-cli-kots-release-app",
					"sequence": int64(1),
				},
			},
		})

	pact.AddInteraction().
		Given("Get KOTS release").
		UponReceiving("A request to get a kots release").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-kots-release-app/release/1"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-kots-release-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"release": map[string]interface{}{
					"appId":    "replicated-cli-kots-release-app",
					"sequence": int64(1),
					"spec":     base64.StdEncoding.EncodeToString([]byte(testYAML)),
				},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_ListReleases(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-list-releases-token")
		client := VendorV3Client{HTTPClient: *api}

		releases, err := client.ListReleases("replicated-cli-list-releases-app")
		assert.NoError(t, err)

		assert.Len(t, releases, 2)

		return nil
	}

	pact.AddInteraction().
		Given("List KOTS releases").
		UponReceiving("A request to list kots releases").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-list-releases-app/releases"),
			Query: dsl.MapMatcher{
				"currentPage": dsl.String("0"),
				"pageSize":    dsl.String("20"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-releases-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"releases": []map[string]interface{}{
					{
						"appId":    "replicated-cli-list-releases-app",
						"sequence": int64(1),
						"spec":     base64.StdEncoding.EncodeToString([]byte("testspec")),
					},
					{
						"appId":    "replicated-cli-list-releases-app",
						"sequence": int64(2),
						"spec":     base64.StdEncoding.EncodeToString([]byte("testspec")),
					},
				},
			},
		})
	pact.AddInteraction().
		Given("List KOTS releases, page 2").
		UponReceiving("A request to list kots releases").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/v3/app/replicated-cli-list-releases-app/releases"),
			Query: dsl.MapMatcher{
				"currentPage": dsl.String("1"),
				"pageSize":    dsl.String("20"),
			},
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-list-releases-token"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: map[string]interface{}{
				"releases": []map[string]interface{}{},
			},
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_UpdateRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-update-release-token")
		client := VendorV3Client{HTTPClient: *api}

		err = client.UpdateRelease("replicated-cli-update-release-app", 1, "testspec2")
		assert.NoError(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Update KOTS release").
		UponReceiving("A request to update a kots release").
		WithRequest(dsl.Request{
			Method: "PUT",
			Path:   dsl.String("/v3/app/replicated-cli-update-release-app/release/1"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-update-release-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"spec_gzip": "H4sIAAAAAAAA/ypJLS4pLkhNNgIEAAD//2X3KIMJAAAA",
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func Test_PromoteRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-promote-release-token")
		client := VendorV3Client{HTTPClient: *api}

		err = client.PromoteRelease("replicated-cli-promote-release-app", 1, "v0.0.1", "releasenotes", false, "replicated-cli-promote-release-unstable")
		assert.NoError(t, err)

		return nil
	}

	pact.AddInteraction().
		Given("Promote KOTS release").
		UponReceiving("A request to promote a kots release").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   dsl.String("/v3/app/replicated-cli-promote-release-app/release/1/promote"),
			Headers: dsl.MapMatcher{
				"Authorization": dsl.String("replicated-cli-promote-release-token"),
				"Content-Type":  dsl.String("application/json"),
			},
			Body: map[string]interface{}{
				"releaseNotes": "releasenotes",
				"versionLabel": "v0.0.1",
				"isRequired":   false,
				"channelIds": []string{
					"replicated-cli-promote-release-unstable",
				},
				"ignoreWarnings": false,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
