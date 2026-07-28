package kotsclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestListChannelReleases_Pagination(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/app/app-id/channel/channel-id/releases", r.URL.Path)
		require.Equal(t, "test-token", r.Header.Get("Authorization"))

		page := r.URL.Query().Get("currentPage")
		pageSize := r.URL.Query().Get("pageSize")
		require.Equal(t, "20", pageSize)
		pageCount += 1

		releases := []types.ChannelRelease{}
		switch page {
		case "0":
			for i := 0; i < 20; i++ {
				releases = append(releases, types.ChannelRelease{Semver: "1.0.0"})
			}
		case "1":
			releases = append(releases, types.ChannelRelease{Semver: "2.17.12"})
		default:
			// Empty page signals end of results
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"releases": releases,
		})
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	releases, err := client.ListChannelReleases("app-id", "channel-id", "")
	require.NoError(t, err)
	require.Len(t, releases, 21)
	require.Equal(t, "2.17.12", releases[20].Semver)
	require.Equal(t, 3, pageCount)
}

func TestListChannelReleases_Pagination_ExactPageSize(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/app/app-id/channel/channel-id/releases", r.URL.Path)

		page := r.URL.Query().Get("currentPage")
		pageCount += 1

		releases := []types.ChannelRelease{}
		switch page {
		case "0", "1":
			for i := 0; i < 20; i++ {
				releases = append(releases, types.ChannelRelease{Semver: "1.0.0"})
			}
		default:
			// Empty page signals end of results
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"releases": releases,
		})
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	releases, err := client.ListChannelReleases("app-id", "channel-id", "")
	require.NoError(t, err)
	require.Len(t, releases, 40)
	require.Equal(t, 3, pageCount)
}

func TestListChannelReleasesByVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/app/app-id/channel/channel-id/releases", r.URL.Path)
		require.Equal(t, "test-token", r.Header.Get("Authorization"))
		require.Equal(t, "2.17.12", r.URL.Query().Get("versionLabel"))
		require.Equal(t, "20", r.URL.Query().Get("pageSize"))

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"releases": []types.ChannelRelease{
				{Semver: "2.17.12"},
			},
		})
	}))
	defer server.Close()

	api := platformclient.NewHTTPClient(server.URL, "test-token")
	client := VendorV3Client{HTTPClient: *api}

	releases, err := client.ListChannelReleasesByVersion("app-id", "channel-id", "2.17.12", "")
	require.NoError(t, err)
	require.Len(t, releases, 1)
	require.Equal(t, "2.17.12", releases[0].Semver)
}
