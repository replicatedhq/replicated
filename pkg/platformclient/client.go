// Package platformclient manages channels and releases through the Replicated Vendor API.
package platformclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/integration"
	kotsclienttypes "github.com/replicatedhq/replicated/pkg/kotsclient/types"
	"github.com/replicatedhq/replicated/pkg/version"
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
	Body       []byte
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

	bodyBytes, err := io.ReadAll(resp.Body)
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

// DoJSON makes the request, and respBody is a pointer to the struct that we should unmarshal the response into
func (c *HTTPClient) DoJSON(ctx context.Context, method string, path string, successStatus int, reqBody interface{}, respBody interface{}) error {
	if ctx.Value(integration.APICallLogContextKey) != nil {
		filename := ctx.Value(integration.APICallLogContextKey).(string)

		// Open the file in append mode, create if it doesn't exist
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening or creating file: %v", err)
		}
		defer f.Close()

		// Format the log entry as METHOD:PATH
		logEntry := fmt.Sprintf("%s:%s\n", strings.ToUpper(method), path)

		// Write the log entry to the file
		if _, err := f.WriteString(logEntry); err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}

		f.Close()
	}
	if ctx.Value(integration.IntegrationTestContextKey) != nil {
		if respBody != nil {
			testResponse := integration.Response(ctx.Value(integration.IntegrationTestContextKey).(string)).(kotsclienttypes.KotsAppResponse)
			encoded, err := json.Marshal(testResponse)
			if err != nil {
				return err
			}
			if err := json.NewDecoder(bytes.NewReader(encoded)).Decode(respBody); err != nil {
				return err
			}
		}

		return nil
	}

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
	req.Header.Set("User-Agent", fmt.Sprintf("Replicated/%s", version.Version()))

	if _, ok := os.LookupEnv("CI"); ok {
		req.Header.Set("X-Replicated-CI", os.Getenv("CI"))
	}

	if err := addGitHubActionsHeaders(req); err != nil {
		return errors.Wrap(err, "add github actions headers")
	}

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
			// look for a response message in the body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return ErrForbidden
			}

			// some of the methods in the api have a standardized response for 403
			type forbiddenResponse struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			var fr forbiddenResponse
			if err := json.Unmarshal(body, &fr); err == nil {
				return errors.New(fr.Error.Message)
			}

			return ErrForbidden
		}
		body, _ := io.ReadAll(resp.Body)
		return APIError{
			Method:     method,
			Endpoint:   endpoint,
			StatusCode: resp.StatusCode,
			Message:    responseBodyToErrorMessage(body),
			Body:       body,
		}
	}
	if respBody != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read body")
		}

		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(respBody); err != nil {
			return fmt.Errorf("%s %s response decoding: %w", method, endpoint, err)
		}
	}

	return nil
}

func addGitHubActionsHeaders(req *http.Request) error {
	// anyone can set this to false to disable this behavior
	if os.Getenv("CI") != "true" {
		return nil
	}

	// the following params are used to link CMX runs back to the workflow
	req.Header.Set("X-Replicated-CI", "true")
	if os.Getenv("GITHUB_RUN_ID") != "" {
		req.Header.Set("X-Replicated-GitHubRunID", os.Getenv("GITHUB_RUN_ID"))
	}
	if os.Getenv("GITHUB_RUN_NUMBER") != "" {
		req.Header.Set("X-Replicated-GitHubRunNumber", os.Getenv("GITHUB_RUN_NUMBER"))
	}
	if os.Getenv("GITHUB_SERVER_URL") != "" {
		req.Header.Set("X-Replicated-GitHubServerURL", os.Getenv("GITHUB_SERVER_URL"))
	}
	if os.Getenv("GITHUB_REPOSITORY") != "" {
		req.Header.Set("X-Replicated-GitHubRepository", os.Getenv("GITHUB_REPOSITORY"))
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s %d: %s", endpoint, resp.StatusCode, body)
	}

	return io.ReadAll(resp.Body)
}

var knownErrorCodes = map[string]string{
	"CUSTOMER_CHANNEL_ERROR": `You are attempting to assign a KOTS-enabled customer to a channel with a Helm-only head release. To resolve this, you can either

1. Disable KOTS installations for this customer by passing the --kots-install=false flag
2. Assign this customer to a different channel that includes a KOTS-capable release
3. Promote a release containing KOTS manifests to this channel (it will still be installable with the Helm CLI as long as it contains a helm chart)`,
}

func responseBodyToErrorMessage(body []byte) string {
	u := map[string]interface{}{}
	if err := json.Unmarshal(body, &u); err != nil {
		return string(body)
	}

	if code, ok := u["error_code"].(string); ok && code != "" {
		m := knownErrorCodes[code]
		if m != "" {
			return m
		}
	}

	if m, ok := u["message"].(string); ok && m != "" {
		return m
	}

	return string(body)
}
