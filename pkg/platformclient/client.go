// Package platformclient manages channels and releases through the Replicated Vendor API.
package platformclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiOrigin = "https://api.replicated.com/vendor"

type AppOptions struct {
	Name string
}

type ChannelOptions struct {
	Name        string
	Description string
}

// An HTTPClient communicates with the Replicated Vendor HTTP API.
// TODO: rename this to client
type HTTPClient struct {
	apiKey    string
	apiOrigin string
}

// New returns a new  HTTP client.
func New(apiKey string) *HTTPClient {
	return NewHTTPClient(apiOrigin, apiKey)
}

func NewHTTPClient(origin string, apiKey string) *HTTPClient {
	c := &HTTPClient{
		apiKey:    apiKey,
		apiOrigin: origin,
	}

	return c
}

func (c *HTTPClient) doJSON(method, path string, successStatus int, reqBody, respBody interface{}) error {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, endpoint, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != successStatus {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s %s %d: %s", method, endpoint, resp.StatusCode, body)
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %v", method, endpoint, err)
		}
	}

	return nil
}
