package kotsclient

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	realkotsclient "github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/assert"
)

func Test_CreateGetRelease(t *testing.T) {
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d", pact.Server.Port)

		api := platformclient.NewHTTPClient(u, "replicated-cli-kots-release-token")
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		releaseInfo, err := client.CreateRelease("replicated-cli-kots-release-app", testMultiYAML)
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
				"spec_gzip": "H4sIAAAAAAAA/7xWS2/jNhD+K4QuvqzMJMUWBoEe2u0laBMYXbSXOAeaHkus+So5dNYN8t8L6i1ZdtAcljfN4+PMNw/q6TUzXEPGMvjGtVOQ78Ape9JgcHniWmWfMsexvGogrEEwmLEsz/ON4U7+BT5IaxjhzgV6vN2YgzQ7Rn7tfDdGA/IdR842hpAUBCOmkOZb+lR8CypUGpIwOlVwICpxAAUCrW9sNEdR/j70GvsRgqCd4gitw/D2dNTYeepeSYyxyFFaM7TbcnGIbnkEBd4upaW1ID9aFTWEBiNvSKr9ujTSSRouDfgeNZ8SUh+peXEurS96sNHgMK4Rxvj6+ujkseZYMkJj8DSU3AOtzGmJWtHe2EOw0QsY4hOipJY4FiVitfUnRhZ3n398kIuxUrjIyOLzzY0eKTz8EyFchvrh7gLS7RCpJXyWxHMChDV7WTxwN7p11iX1eCnVzoPJ2NPz26ezqQngj1LAhZGZaPt5GY5KPyVfa/MPjshcnx4shtSaLyUYRhavrx6cIsYiuQ+/Ra/I29tiMFx4csDIFxUDgr9fJ5GzvqtPXn0xsrqZmcRhKGfr4PvneCW/R7uDtfV4Pb2qKRpLRlar1epq0pdbJcWXc+cmPTIVX1mmbYbH2y0g76j82TklRcXFNTr77CUqqNzII9eQRFIk/BLRBUap5y/LQmIZtzGAbwJaCqupMGJPuccX6w9U89Qe1Hn7NwgM9BC34A0gBJrwqLDK+oE0T9K8ki6dKSoakWMM92Zvve73X076N4Z2FR+XqBmpxyrDTVZbZW3BGu16UkdlBVfjQjbVawn806sEl5hglHao71S1XiVzhR1r/n9tv1T+s2UVjaqra+FtdIGRp+d3wnUe9koW5fSFn1fObyv0Nm4VhNJaXIayifuujXvdosyG3t2RixLEIQyS4Iar07+pF97PI0SXmiLfRrNT0917xeJDGX2toX6pkGazGl82yElYVW+Lvn1Fs1vN3jLy+jYR/9E9twOdssXwgZxsoPbJ4879NPk/SNEFxwX0W/GxFdWLccrz838BAAD//22E3lIXCgAA",
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
					"spec":     testMultiYAML,
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

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
						"sequence": int64(2),
						"spec":     "",
					},
					{
						"appId":    "replicated-cli-list-releases-app",
						"sequence": int64(1),
						"spec":     "",
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

		err = client.UpdateRelease("replicated-cli-update-release-app", 1, testMultiYAML)
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
				"spec_gzip": "H4sIAAAAAAAA/7xWS2/jNhD+K4QuvqzMJMUWBoEe2u0laBMYXbSXOAeaHkus+So5dNYN8t8L6i1ZdtAcljfN4+PMNw/q6TUzXEPGMvjGtVOQ78Ape9JgcHniWmWfMsexvGogrEEwmLEsz/ON4U7+BT5IaxjhzgV6vN2YgzQ7Rn7tfDdGA/IdR842hpAUBCOmkOZb+lR8CypUGpIwOlVwICpxAAUCrW9sNEdR/j70GvsRgqCd4gitw/D2dNTYeepeSYyxyFFaM7TbcnGIbnkEBd4upaW1ID9aFTWEBiNvSKr9ujTSSRouDfgeNZ8SUh+peXEurS96sNHgMK4Rxvj6+ujkseZYMkJj8DSU3AOtzGmJWtHe2EOw0QsY4hOipJY4FiVitfUnRhZ3n398kIuxUrjIyOLzzY0eKTz8EyFchvrh7gLS7RCpJXyWxHMChDV7WTxwN7p11iX1eCnVzoPJ2NPz26ezqQngj1LAhZGZaPt5GY5KPyVfa/MPjshcnx4shtSaLyUYRhavrx6cIsYiuQ+/Ra/I29tiMFx4csDIFxUDgr9fJ5GzvqtPXn0xsrqZmcRhKGfr4PvneCW/R7uDtfV4Pb2qKRpLRlar1epq0pdbJcWXc+cmPTIVX1mmbYbH2y0g76j82TklRcXFNTr77CUqqNzII9eQRFIk/BLRBUap5y/LQmIZtzGAbwJaCqupMGJPuccX6w9U89Qe1Hn7NwgM9BC34A0gBJrwqLDK+oE0T9K8ki6dKSoakWMM92Zvve73X076N4Z2FR+XqBmpxyrDTVZbZW3BGu16UkdlBVfjQjbVawn806sEl5hglHao71S1XiVzhR1r/n9tv1T+s2UVjaqra+FtdIGRp+d3wnUe9koW5fSFn1fObyv0Nm4VhNJaXIayifuujXvdosyG3t2RixLEIQyS4Iar07+pF97PI0SXmiLfRrNT0917xeJDGX2toX6pkGazGl82yElYVW+Lvn1Fs1vN3jLy+jYR/9E9twOdssXwgZxsoPbJ4879NPk/SNEFxwX0W/GxFdWLccrz838BAAD//22E3lIXCgAA",
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
		client := realkotsclient.VendorV3Client{HTTPClient: *api}

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
				"ignoreWarnings":        false,
				"omitDetailsInResponse": true,
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	if err := pact.Verify(test); err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
