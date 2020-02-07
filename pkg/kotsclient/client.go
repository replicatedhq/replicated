package kotsclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"io/ioutil"
	"net/http"
)

var ErrNotFound = errors.New("Not found")

type AppOptions struct {
	Name string
}

type ChannelOptions struct {
	Name        string
	Description string
}

// Client communicates with the Replicated Vendor GraphQL API.
type HybridClient struct {
	// GraphQL client for g.replicated.com (graphql-api)
	GraphQLClient *graphql.Client

	// for rest-y KOTS operations against g.replicated.com
	httpOrigin string
	apiKey     string
}

func NewHybridClient(graphqlOrigin string, apiKey string, httpOrigin string) *HybridClient {
	c := &HybridClient{
		GraphQLClient: graphql.NewClient(graphqlOrigin, apiKey),
		httpOrigin:    httpOrigin,
		apiKey:        apiKey,
	}

	return c
}

func (c *HybridClient) ExecuteGraphQLRequest(requestObj graphql.Request, deserializeTarget interface{}) error {
	return c.GraphQLClient.ExecuteRequest(requestObj, deserializeTarget)
}

func (c *HybridClient) ExecuteHTTPRequest(requestObj graphql.Request, deserializeTarget interface{}) error {
	return c.GraphQLClient.ExecuteRequest(requestObj, deserializeTarget)
}

func (c *HybridClient) doRawHTTP(method, path string, successStatus int, reqBody interface{}) ([]byte, error) {
	endpoint := fmt.Sprintf("%s%s", c.httpOrigin, path)
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, endpoint, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != successStatus {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s %s %d: %s", method, endpoint, resp.StatusCode, body)
	}

	return ioutil.ReadAll(resp.Body)
}
