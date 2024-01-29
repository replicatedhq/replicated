package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type Client struct {
	PlatformClient *platformclient.HTTPClient
	KotsClient     *kotsclient.VendorV3Client
}

func NewClient(platformOrigin string, apiToken string, kurlOrigin string) Client {
	httpClient := platformclient.NewHTTPClient(platformOrigin, apiToken)
	client := Client{
		PlatformClient: httpClient,
		KotsClient:     &kotsclient.VendorV3Client{HTTPClient: *httpClient},
	}

	return client
}

func (c *Client) GetAppType(appID string) (*types.App, string, error) {
	platformSwaggerApp, err1 := c.PlatformClient.GetApp(appID)
	if err1 == nil && platformSwaggerApp != nil {
		platformApp := &types.App{
			ID:        platformSwaggerApp.Id,
			Name:      platformSwaggerApp.Name,
			Slug:      platformSwaggerApp.Slug,
			Scheduler: platformSwaggerApp.Scheduler,
		}
		return platformApp, "platform", nil
	}

	kotsApp, err2 := c.KotsClient.GetApp(appID)
	if err2 == nil && kotsApp != nil {
		return kotsApp, "kots", nil
	}

	err := errors.Errorf("Following errors occurred while trying to get app type: error 1: %s, error 2: %s", err1, err2)
	return nil, "", err
}
