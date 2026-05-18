package client

import (
	"testing"

	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/stretchr/testify/require"
)

func newTestClient(origin string) Client {
	platformHTTPClient := platformclient.NewHTTPClient(origin, "fake-api-key")
	return Client{
		PlatformClient: platformHTTPClient,
		KotsClient:     &kotsclient.VendorV3Client{HTTPClient: *platformHTTPClient},
	}
}

func TestClientGetAppReturnsBackendError(t *testing.T) {
	c := newTestClient("http://%")

	app, err := c.GetApp("app-id")
	require.Nil(t, app)
	require.Error(t, err)
}

func TestClientCreateAppRejectsUnsupportedOptions(t *testing.T) {
	c := newTestClient("http://127.0.0.1")

	app, err := c.CreateApp(struct{}{})
	require.Nil(t, app)
	require.ErrorContains(t, err, "unsupported app options type")
}

func TestClientCreateAppRejectsNilOptions(t *testing.T) {
	c := newTestClient("http://127.0.0.1")

	app, err := c.CreateApp((*platformclient.AppOptions)(nil))
	require.Nil(t, app)
	require.ErrorContains(t, err, "create app options cannot be nil")
}

func TestClientDeleteAppReturnsLookupError(t *testing.T) {
	c := newTestClient("http://%")

	require.Error(t, c.DeleteApp("app-id"))
}
