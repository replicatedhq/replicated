package shipclient

import (
	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type Client interface {
	ListApps() ([]types.AppAndChannels, error)
	GetApp(appID string) (*types.App, error)

	ListChannels(string) ([]types.Channel, error)
	CreateChannel(string, string, string) error

	ListReleases(appID string) ([]types.ReleaseInfo, error)
	CreateRelease(appID string, yaml string) (*types.ReleaseInfo, error)
	UpdateRelease(appID string, sequence int64, yaml string) error
	PromoteRelease(appID string, sequence int64, label string, notes string, channelIDs ...string) error
	LintRelease(string, string) ([]types.LintMessage, error)

	ListCollectors(appID string, appType string) ([]v1.AppCollectorInfo, error)
	CreateCollector(appID string, appType string, name string, yaml string) (*v1.AppCollectorInfo, error)
	UpdateCollector(appID string, appType string, specID string, yaml string) (interface{}, error)
	UpdateCollectorName(appID string, appType string, specID, name string) (interface{}, error)
	GetCollector(appID string, appType string, specID string) (*v1.AppCollectorInfo, error)
	PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error

	CreateEntitlementSpec(appID string, name string, spec string) (*types.EntitlementSpec, error)
	SetDefaultEntitlementSpec(specID string) error
	SetEntitlementValue(customerID string, specID string, key string, value string, datatype string, appID string) (*types.EntitlementValue, error)
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
