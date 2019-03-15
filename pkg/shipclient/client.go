package shipclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/replicatedhq/replicated/pkg/types"
)

const apiOrigin = "https://g.replicated.com/graphql"

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
}

type AppOptions struct {
	Name string
}

type ChannelOptions struct {
	Name        string
	Description string
}

// GraphQLClient communicates with the Replicated Vendor GraphQL API.
type GraphQLClient struct {
	GQLServer *url.URL
	Token     string
}

func NewGraphQLClient(origin string, apiKey string) Client {
	uri, err := url.Parse(origin)
	if err != nil {
		panic(err)
	}

	c := &GraphQLClient{
		GQLServer: uri,
		Token:     apiKey,
	}

	return c
}

// GraphQLRequest is a json-serializable request to the graphql server
type GraphQLRequest struct {
	Query         string                 `json:"query,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// GraphQLError represents an error returned by the graphql server
type GraphQLError struct {
	Locations []map[string]interface{} `json:"locations"`
	Message   string                   `json:"message"`
	Code      string                   `json:"code"`
}

type GraphQLResponseErrorOnly struct {
	Errors []GraphQLError `json:"errors,omitempty"`
}

type ShipError interface {
	GraphQLError() []GraphQLError
}

func (c *GraphQLClient) executeRequest(requestObj GraphQLRequest, deserializeTarget interface{}) error {
	body, err := json.Marshal(requestObj)
	if err != nil {
		return err
	}

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", c.GQLServer.String(), bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp == nil {
		return err
	}

	if resp.Body == nil {
		return err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, deserializeTarget); err != nil {
		return err
	}

	return nil
}

func (c *GraphQLClient) checkErrors(shipError ShipError) error {
	if shipError.GraphQLError() != nil && len(shipError.GraphQLError()) > 0 {
		var multiErr *multierror.Error
		for _, err := range shipError.GraphQLError() {
			multiErr = multierror.Append(multiErr, fmt.Errorf("%s: %s", err.Code, err.Message))

		}
		return multiErr.ErrorOrNil()
	}
	return nil
}
