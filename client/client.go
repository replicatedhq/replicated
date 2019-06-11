package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/shipclient"
)

type Client struct {
	PlatformClient platformclient.Client
	ShipClient     shipclient.Client
}

func NewClient(platformOrigin string, graphqlOrigin string, apiToken string) Client {
	client := Client{
		PlatformClient: platformclient.NewHTTPClient(platformOrigin, graphqlOrigin, apiToken),
		ShipClient:     shipclient.NewGraphQLClient(graphqlOrigin, apiToken),
	}

	return client
}

func (c *Client) GetAppType(appID string) (string, error) {
	platformApp, err := c.PlatformClient.GetApp(appID)
	if err == nil && platformApp != nil {
		return "platform", nil
	}

	shipApp, err := c.ShipClient.GetApp(appID)
	if err == nil && shipApp != nil {
		return "ship", nil
	}

	return "", errors.New("app not found")
}
