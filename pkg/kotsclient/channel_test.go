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
