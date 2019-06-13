package platformclient

import (
	"fmt"
	"net/http"
	"time"

	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

type GraphQLResponseListCollectors struct {
	Data *SupportBundleSpecsData `json:"data,omitempty"`
}

type GraphQLResponseUpdateCollector struct {
	Data *SupportBundleUpdateSpecData `json:"data,omitempty"`
}

type SupportBundleUpdateSpecData struct {
	UpdateSupportBundleSpec *UpdateSupportBundleSpec `json:"updateSupportBundleSpec"`
}

type UpdateSupportBundleSpec struct {
	ID     string `json:"id"`
	Config string `json:"spec,omitempty"`
}

type GraphQLResponseGetCollector struct {
	Data *SupportBundleGetSpecData `json:"data,omitempty"`
}

type SupportBundleGetSpecData struct {
	GetSupportBundleSpec *GetSupportBundleSpec `json:"supportBundleSpec"`
}

type GetSupportBundleSpec struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Config    string `json:"spec,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type GraphQLResponseUpdateNameCollector struct {
	Data *SupportBundleUpdateSpecNameData `json:"data,omitempty"`
}

type SupportBundleUpdateSpecNameData struct {
	UpdateSupportBundleSpecName *UpdateSupportBundleSpecName `json:"updateSupportBundleSpecName"`
}

type UpdateSupportBundleSpecName struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type GraphQLResponseCreateCollector struct {
	Data *SupportBundleCreateSpec `json:"data,omitempty"`
}

type SupportBundleCreateSpec struct {
	CreateSupportBundleSpec *CreateSupportBundleSpec `json:"createSupportBundleSpec,omitempty"`
}

type CreateSupportBundleSpec struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
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

func (c *HTTPClient) ListCollectors(appID string, appType string) ([]v1.AppCollectorInfo, error) {
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
func (c *HTTPClient) GetCollector(appID string, appType string, id string) (*v1.AppCollectorInfo, error) {
	response := GraphQLResponseGetCollector{}

	request := graphql.Request{
		Query: `
query supportBundleSpec($id: String!) {
	supportBundleSpec(id: $id) {
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
	}
}
`,

		Variables: map[string]interface{}{
			"id": id,
		},
	}

	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	getCollectorInfo := v1.AppCollectorInfo{
		AppId:  appID,
		SpecId: response.Data.GetSupportBundleSpec.ID,
		Config: response.Data.GetSupportBundleSpec.Config,
		Name:   response.Data.GetSupportBundleSpec.Name,
	}

	return &getCollectorInfo, nil
}

func (c *HTTPClient) UpdateCollector(appID string, appType string, specID, yaml string) (interface{}, error) {
	response := GraphQLResponseUpdateCollector{}

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
		return nil, err
	}

	return &response, nil
}

// Vendor-API: PromoteCollector points the specified channels at a named collector.
func (c *HTTPClient) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {
	path := fmt.Sprintf("/v1/app/%s/collector/%s/promote", appID, specID)
	body := &v1.BodyPromoteCollector{
		ChannelIDs: channelIDs,
	}
	if err := c.doJSON("POST", path, http.StatusOK, body, nil); err != nil {
		return fmt.Errorf("PromoteCollector: %v", err)
	}
	return nil
}

// CreateCollector - input appID, name, yaml - return Name, Spec, Config
func (c *HTTPClient) CreateCollector(appID string, appType string, name string, yaml string) (*v1.AppCollectorInfo, error) {
	response := GraphQLResponseCreateCollector{}

	request := graphql.Request{
		Query: `
mutation createSupportBundleSpec($name: String, $appId: String, $spec: String, $githubRef: GitHubRefInput) {
	createSupportBundleSpec(name: $name, appId: $appId, spec: $spec, githubRef: $githubRef) {
		id
		name
		spec
		createdAt
		updatedAt
		githubRef {
		owner
		repoFullName
		branch
		path
		}
	}
}`,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
			"name":  name,
		},
	}

	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	request = graphql.Request{
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
			"id":   response.Data.CreateSupportBundleSpec.ID,
			"spec": yaml,
		},
	}

	finalizeSpecCreate := GraphQLResponseUpdateCollector{}
	if err := c.graphqlClient.ExecuteRequest(request, &finalizeSpecCreate); err != nil {
		return nil, err
	}

	newCollectorInfo := v1.AppCollectorInfo{
		AppId:  appID,
		SpecId: finalizeSpecCreate.Data.UpdateSupportBundleSpec.ID,
		Config: finalizeSpecCreate.Data.UpdateSupportBundleSpec.Config,
		Name:   response.Data.CreateSupportBundleSpec.Name,
	}

	return &newCollectorInfo, nil

}

func (c *HTTPClient) UpdateCollectorName(appID string, appType string, specID, name string) (interface{}, error) {
	response := GraphQLResponseUpdateNameCollector{}

	request := graphql.Request{
		Query: `
mutation updateSupportBundleSpecName($id: ID!, $name: String!) {
	updateSupportBundleSpecName(id: $id, name: $name) {
	id
	name
	}
}
`,

		Variables: map[string]interface{}{
			"id":   specID,
			"name": name,
		},
	}

	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
