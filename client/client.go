// Package client manages channels and releases through the Replicated Vendor API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// A Client communicates with the Replicated Vendor API.
type Client struct {
	apiKey    string
	apiOrigin string
}

// New returns a new client.
func New(origin string, apiKey string) *Client {
	c := &Client{
		apiKey:    apiKey,
		apiOrigin: origin,
	}

	return c
}

func (c *Client) doJSON(method, path string, successStatus int, reqBody, respBody interface{}) error {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return fmt.Errorf("%s %s: %v", method, endpoint, err)
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
		return fmt.Errorf("%s %s: %v", method, endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != successStatus {
		return fmt.Errorf("%s %s: status %d", method, endpoint, resp.StatusCode)
	}
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %v", method, endpoint, err)
		}
	}
	return nil
}
