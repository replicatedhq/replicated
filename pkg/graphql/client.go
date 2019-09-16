package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	multierror "github.com/hashicorp/go-multierror"
)

const APIOrigin = "https://g.replicated.com/graphql"

// Client communicates with the Replicated Vendor GraphQL API.
type Client struct {
	GQLServer *url.URL
	Token     string
}

func NewClient(origin string, apiKey string) *Client {
	uri, err := url.Parse(origin)
	if err != nil {
		panic(err)
	}

	c := &Client{
		GQLServer: uri,
		Token:     apiKey,
	}

	return c
}

// Request is a json-serializable request to the graphql server
type Request struct {
	Query         string                 `json:"query,omitempty"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// Error represents an error returned by the graphql server
type Error struct {
	Locations []map[string]interface{} `json:"locations"`
	Message   string                   `json:"message"`
	Code      string                   `json:"code"`
}

type GQLError interface {
	GraphQLError() []Error
}

type ResponseErrorOnly struct {
	Errors []Error `json:"errors,omitempty"`
}

func (r ResponseErrorOnly) GraphQLError() []Error {
	return r.Errors
}

func (c *Client) ExecuteRequest(requestObj Request, deserializeTarget interface{}) error {
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

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var gqlErr ResponseErrorOnly
	_ = json.Unmarshal(responseBody, &gqlErr) // ignore error to be safe

	if err := c.checkErrors(gqlErr); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	if err := json.Unmarshal(responseBody, deserializeTarget); err != nil {
		return err
	}

	return nil
}

func (c *Client) checkErrors(gqlError ResponseErrorOnly) error {
	if len(gqlError.GraphQLError()) > 0 {
		var multiErr *multierror.Error
		for _, err := range gqlError.GraphQLError() {
			multiErr = multierror.Append(multiErr, fmt.Errorf("%s: %s", err.Code, err.Message))
		}
		return multiErr.ErrorOrNil()
	}
	return nil
}
