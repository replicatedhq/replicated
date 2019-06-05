package platformclient

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

type GraphQLResponseListCollectors struct {
	Data   *SupportBundleSpecsData `json:"data,omitempty"`
	Errors []graphql.GQLError      `json:"errors,omitempty"`
}

type SupportBundleSpecsData struct {
	SupportBundleSpecs []SupportBundleSpec `json:"supportBundleSpecs"`
}

type SupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []v1.AppChannel `json:"platformChannels"`
}

type PlatformChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *HTTPClient) ListCollectors(appID string) ([]v1.AppCollectorInfo, error) {
	response := GraphQLResponseListCollectors{}

	request := graphql.Request{
		Query: `
query supportBundleSpecs($appId: String) {
  supportBundleSpecs(appId: $appId) {
    id
    name
    spec
    createdAt
    updatedAt
    isArchived
    isImmutable
    githubRef {
      owner
      repoFullName
      branch
      path
    }
    channels {
      id
      name
    }
    platformChannels {
      id
      name
    }
  }
}`,

		Variables: map[string]interface{}{
			"appId": appID,
		},
	}

	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	collectors := make([]v1.AppCollectorInfo, 0, 0)
	for _, spec := range response.Data.SupportBundleSpecs {
		createdAt, err := time.Parse(time.RFC3339, spec.CreatedAt)
		if err != nil {
			return nil, err
		}
		collector := v1.AppCollectorInfo{
			SpecId:         spec.ID,
			Name:           spec.Name,
			CreatedAt:      createdAt.In(location),
			ActiveChannels: spec.Channels,
		}

		collectors = append(collectors, collector)
	}

	return collectors, nil
}

// CreateCollector adds a release to an app.
func (c *HTTPClient) CreateCollector(appID string, name string, yaml string) (*v1.AppCollectorInfo, error) {
	path := fmt.Sprintf("/v1/app/%s/collector/%d", appID, name)
	body := &v1.BodyCreateCollector{
		Source: "latest",
	}
	collector := &v1.AppCollectorInfo{}
	if err := c.doJSON("POST", path, http.StatusCreated, body, collector); err != nil {
		return nil, fmt.Errorf("CreateCollector: %v", err)
	}
	// API does not accept yaml in create operation, so first create then udpate
	if yaml != "" {
		if err := c.UpdateCollector(appID, collector.Name, yaml); err != nil {
			return nil, fmt.Errorf("CreateCollector with YAML: %v", err)
		}
	}
	return collector, nil
}

// UpdateCollector updates a collector's yaml.
func (c *HTTPClient) UpdateCollector(appID string, name string, yaml string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s/collectors/%d/raw", c.apiOrigin, appID, name)
	req, err := http.NewRequest("PUT", endpoint, strings.NewReader(yaml))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/yaml")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateCollector: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if badRequestErr, err := unmarshalBadRequest(resp.Body); err == nil {
			return badRequestErr
		}
		return fmt.Errorf("UpdateRelease (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}

// GetCollector returns a collector's properties.
func (c *HTTPClient) GetCollector(appID string, name string) (*v1.AppCollector, error) {
	path := fmt.Sprintf("/v1/app/%s/collectors/%d/properties", appID, name)
	collector := &v1.AppCollector{}
	if err := c.doJSON("GET", path, http.StatusOK, nil, collector); err != nil {
		return nil, fmt.Errorf("GetCollector: %v", err)
	}
	return collector, nil
}

// PromoteCollector points the specified channels at a named collector.
func (c *HTTPClient) PromoteCollector(appID string, name string, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/collectors/%d/promote?dry_run=true", appID, name)
	body := &v1.BodyPromoteCollector{
		Channels: channelIDs,
	}
	if err := c.doJSON("POST", path, http.StatusNoContent, body, nil); err != nil {
		return fmt.Errorf("PromoteCollector: %v", err)
	}
	return nil
}
