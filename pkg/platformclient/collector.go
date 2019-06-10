package platformclient

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

type GraphQLResponseListCollectors struct {
	Data   *SupportBundleSpecsData `json:"data,omitempty"`
	Errors []graphql.GQLError      `json:"errors,omitempty"`
}

type GraphQLResponseUpdateCollector struct {
	Data *SupportBundleUpdateSpecData `json:"data,omitempty"`
	// Errors []graphql.GQLError           `json:"errors,omitempty"`
}

type SupportBundleUpdateSpecData struct {
	UpdateSupportBundleSpec *UpdateSupportBundleSpec `json:"updateSupportBundleSpec"`
}

type UpdateSupportBundleSpec struct {
	ID     string `json:"id"`
	Config string `json:"spec,omitempty"`
}

type SupportBundleSpecsData struct {
	SupportBundleSpecs []SupportBundleSpec `json:"supportBundleSpecs"`
}

type SupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []v1.AppChannel `json:"platformChannels"`
	Config    string          `json:"spec,omitempty"`
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
			Config:         spec.Config,
		}

		collectors = append(collectors, collector)
	}

	return collectors, nil
}

// GetCollector returns a collector's properties.
func (c *HTTPClient) GetCollector(appID string, id string) (*v1.AppCollectorInfo, error) {
	allcollectors, err := c.ListCollectors(appID)
	if err != nil {
		return nil, err
	}

	for _, collector := range allcollectors {
		if collector.SpecId == id {
			return &collector, nil
		}
	}

	return nil, errors.New("Not found")
}

func (c *HTTPClient) UpdateCollector(appID string, specID, yaml string) error {
	// response := GraphQLResponseUpdateCollector{}
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: `
		mutation updateSupportBundleSpec($id: ID!, $spec: String!, $githubRef: GitHubRefInput, $isArchived: Boolean) {
			updateSupportBundleSpec(id: $id, spec: $spec, githubRef: $githubRef, isArchived: $isArchived) {
				id
				spec
				createdAt
				updatedAt
				isArchived
				githubRef {
					owner
					repoFullName
					branch
					path
				}
			}
		}
	`,

		Variables: map[string]interface{}{
			// "githubRef":  null,
			"id": specID,
			// "isArchived": null,
			"spec": yaml,
		},
	}

	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil
}

// or GQL?, but VenWeb not using it, soooo.....
// PromoteCollector points the specified channels at a named collector.
func (c *HTTPClient) PromoteCollector(appID string, specID string, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/collector/%s/promote", appID, specID)
	body := &v1.BodyPromoteCollector{
		ChannelIDs: channelIDs,
	}
	if err := c.doJSON("POST", path, http.StatusOK, body, nil); err != nil {
		return fmt.Errorf("PromoteCollector: %v", err)
	}
	return nil
}

// TODO: CreateCollector adds a release to an app.
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
