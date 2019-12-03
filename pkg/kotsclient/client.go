package kotsclient

import (
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type Client interface {
	ListApps() ([]types.AppAndChannels, error)
	GetApp(appID string) (*types.App, error)

	CreateRelease(appID string, multiyaml string) (*types.ReleaseInfo, error)
	ListReleases(appID string) ([]types.ReleaseInfo, error)
	UpdateRelease(appID string, sequence int64, yaml string) error
	PromoteRelease(appID string, sequence int64, label string, notes string, channelIDs ...string) error

	ListChannels(appID string) ([]types.Channel, error)
	CreateChannel(appID string, name string, description string) (string, error)
	GetChannel(appID, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error)

	ListCustomers(appID string) ([]types.Customer, error)
}

type AppOptions struct {
	Name string
}

type ChannelOptions struct {
	Name        string
	Description string
}

// Client communicates with the Replicated Vendor GraphQL API.
type GraphQLClient struct {
	GraphQLClient *graphql.Client
}


func NewGraphQLClient(origin string, apiKey string) Client {
	c := &GraphQLClient{GraphQLClient: graphql.NewClient(origin, apiKey)}

	return c
}

func (c *GraphQLClient) ExecuteRequest(requestObj graphql.Request, deserializeTarget interface{}) error {
	return c.GraphQLClient.ExecuteRequest(requestObj, deserializeTarget)
}
