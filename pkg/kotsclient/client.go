package kotsclient

import (
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

type AppOptions struct {
	Name string
}

type ChannelOptions struct {
	Name        string
	Description string
}

// Client communicates with the Replicated Vendor GraphQL API.
type GraphQLClient struct {
	GraphQLClient    *graphql.Client
	KurlDotSHAddress string
}

func NewGraphQLClient(origin string, apiKey string, kurlDotSHAddress string) *GraphQLClient {
	c := &GraphQLClient{
		GraphQLClient:    graphql.NewClient(origin, apiKey),
		KurlDotSHAddress: kurlDotSHAddress,
	}

	return c
}

func (c *GraphQLClient) ExecuteRequest(requestObj graphql.Request, deserializeTarget interface{}) error {
	return c.GraphQLClient.ExecuteRequest(requestObj, deserializeTarget)
}

// Putting a wrapper in the kotsclient package for kots-specific methods but
// don't want to re-invent or duplicate  all that logic for initialization,
// instantiation, and the DoJSON method
//
// we should think more about how we want to organize these going forward, but
// I'm inclined to wait until after everything has been moved off of GQL before deciding
type HTTPClient struct {
	platformclient.HTTPClient
}

