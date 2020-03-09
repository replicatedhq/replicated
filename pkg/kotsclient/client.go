package kotsclient

import (
	"github.com/replicatedhq/replicated/pkg/graphql"
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
