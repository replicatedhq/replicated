// Package platformclient manages channels and releases through the Replicated Vendor API.
package platformclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	apps "github.com/replicatedhq/replicated/gen/go/v1"
	channels "github.com/replicatedhq/replicated/gen/go/v1"
	collectors "github.com/replicatedhq/replicated/gen/go/v1"
	releases "github.com/replicatedhq/replicated/gen/go/v1"
	v2 "github.com/replicatedhq/replicated/gen/go/v2"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

const apiOrigin = "https://api.replicated.com/vendor"
const graphqlApiOrigin = "https://pg.replicated.com/graphql"

// Client provides methods to manage apps, channels, and releases.
type Client interface {
	ListApps() ([]apps.AppAndChannels, error)
	GetApp(string) (*apps.App, error)
	CreateApp(opts *AppOptions) (*apps.App, error)

	ListChannels(string) ([]channels.AppChannel, error)
	CreateChannel(string, string, string) error
	ArchiveChannel(appID, channelID string) error
	GetChannel(appID, channelID string) (*channels.AppChannel, []channels.ChannelRelease, error)

	ListReleases(appID string) ([]releases.AppReleaseInfo, error)
	CreateRelease(appID string, yaml string) (*releases.AppReleaseInfo, error)
	UpdateRelease(appID string, sequence int64, yaml string) error
	GetRelease(appID string, sequence int64) (*releases.AppRelease, error)
	PromoteRelease(appID string, sequence int64, label string, notes string, required bool, channelIDs ...string) error
	LintRelease(string, string) ([]types.LintMessage, error)

	ListCollectors(appID string) ([]collectors.AppCollectorInfo, error)
	CreateCollector(appID string, name string, yaml string) (*collectors.AppCollectorInfo, error)
	UpdateCollector(appID string, name string, yaml string) error
	GetCollector(appID string, name string) (*collectors.AppCollector, error)
	PromoteCollector(appID string, name string, channelIDs ...string) error

	CreateLicense(*v2.LicenseV2) (*v2.LicenseV2, error)
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

	graphqlClient *graphql.Client
}

// New returns a new  HTTP client.
func New(apiKey string) Client {
	return NewHTTPClient(apiOrigin, graphqlApiOrigin, apiKey)
}

func NewHTTPClient(origin, graphqlOrigin, apiKey string) Client {
	c := &HTTPClient{
		apiKey:        apiKey,
		apiOrigin:     origin,
		graphqlClient: graphql.NewClient(graphqlOrigin, apiKey),
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
