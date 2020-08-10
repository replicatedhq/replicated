// Package platformclient manages channels and releases through the Replicated Vendor API.
package platformclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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

func (c *HTTPClient) DoJSON(method, path string, successStatus int, reqBody, respBody interface{}) error {
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
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read body")
		}
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %v", method, endpoint, err)
		}
	}

	return nil
}


func (c *HTTPClient) HTTPGet(path string, successStatus int) ([]byte, error) {

	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf bytes.Buffer
	req, err := http.NewRequest("GET", endpoint, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apiKey)
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
		return nil, fmt.Errorf("GET %s %d: %s", endpoint, resp.StatusCode, body)
	}

	return ioutil.ReadAll(resp.Body)
}
