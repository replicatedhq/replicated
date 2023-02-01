package client

import (
	"github.com/replicatedhq/replicated/pkg/instancesclient"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/shipclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type Client struct {
	PlatformClient  *platformclient.HTTPClient
	ShipClient      *shipclient.GraphQLClient
	KotsClient      *kotsclient.VendorV3Client
	InstancesClient *instancesclient.InstancesClient
}

func NewClient(platformOrigin string, graphqlOrigin string, apiToken string, kurlOrigin string, unifiedAPIOrigin string) Client {
	httpClient := platformclient.NewHTTPClient(platformOrigin, apiToken)
	unifiedClient := platformclient.NewHTTPClient(unifiedAPIOrigin, apiToken)
	client := Client{
		PlatformClient:  httpClient,
		ShipClient:      shipclient.NewGraphQLClient(graphqlOrigin, apiToken),
		KotsClient:      &kotsclient.VendorV3Client{HTTPClient: *httpClient},
		InstancesClient: &instancesclient.InstancesClient{HTTPClient: *unifiedClient},
	}

	return client
}

func (c *Client) GetAppType(appID string) (*types.App, string, error) {
	platformSwaggerApp, err := c.PlatformClient.GetApp(appID)
	if err == nil && platformSwaggerApp != nil {
		platformApp := &types.App{
			ID:        platformSwaggerApp.Id,
			Name:      platformSwaggerApp.Name,
			Slug:      platformSwaggerApp.Slug,
			Scheduler: platformSwaggerApp.Scheduler,
		}
		return platformApp, "platform", nil
	}

	shipApp, err := c.ShipClient.GetApp(appID)
	if err == nil && shipApp != nil {
		return shipApp, "ship", nil
	}

	kotsApp, err := c.KotsClient.GetApp(appID)
	if err == nil && kotsApp != nil {
		return kotsApp, "kots", nil
	}

	return nil, "", err
}
