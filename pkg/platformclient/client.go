// Package platformclient manages channels and releases through the Replicated Vendor API.
package platformclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const apiOrigin = "https://api.replicated.com/vendor"

var (
	ErrForbidden = errors.New("the action is not allowed for the current user or team")
)

type APIError struct {
	Method     string
	Endpoint   string
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s %s %d: %s", e.Method, e.Endpoint, e.StatusCode, e.Message)
}

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

func (c *HTTPClient) DoJSONWithoutUnmarshal(method string, path string, reqBody string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)
	var buf *bytes.Buffer
	if reqBody != "" {
		buf = bytes.NewBuffer([]byte(reqBody))
	} else {
		buf = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, endpoint, buf)
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	// if the response code was NOT a 2xx code, then we either return what the server responded with
	// or a static error if the server didn't respond with a body
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(bodyBytes) > 0 {
			return nil, errors.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil, errors.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return bodyBytes, nil
}

func (c *HTTPClient) DoJSON(method string, path string, successStatus int, reqBody interface{}, respBody interface{}) error {
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
		if resp.StatusCode == http.StatusForbidden {
			return ErrForbidden
		}
		body, _ := ioutil.ReadAll(resp.Body)
		return APIError{
			Method:     method,
			Endpoint:   endpoint,
			StatusCode: resp.StatusCode,
			Message:    responseBodyToErrorMessage(body),
		}
	}
	if respBody != nil {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read body")
		}

		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %w", method, endpoint, err)
		}
	}

	return nil
}

// Minimal, simplified version of DoJSON for GET requests, just returns bytes
func (c *HTTPClient) HTTPGet(path string, successStatus int) ([]byte, error) {

	endpoint := fmt.Sprintf("%s%s", c.apiOrigin, path)

	req, err := http.NewRequest("GET", endpoint, nil)
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

func responseBodyToErrorMessage(body []byte) string {
	u := map[string]interface{}{}
	if err := json.Unmarshal(body, &u); err != nil {
		return string(body)
	}

	if m, ok := u["message"].(string); ok && m != "" {
		return m
	}

	return string(body)
}
