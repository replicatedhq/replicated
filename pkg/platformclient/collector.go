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

type GraphQLResponseCreateCollector struct {
	Data   *SupportBundleCreateSpec `json:"data,omitempty"`
	Errors []graphql.GQLError       `json:"errors,omitempty"`
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

func (c *HTTPClient) UpdateCollector(appID string, specID, yaml string) (interface{}, error) {
	response := GraphQLResponseUpdateCollector{}
	// response := graphql.ResponseErrorOnly{}

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

// // GraphQL: PromoteCollector assigns collector to a specified channel.
// func (c *HTTPClient) PromoteCollector(appID string, specID string, channelIDs ...string) error {
// 	response := graphql.ResponseErrorOnly{}

// 	request := graphql.Request{
// 		Query: `
// mutation  promoteTroubleshootSpec($channelIds: [String], $specId: ID!) {
// 	promoteTroubleshootSpec(channelIds: $channelIds, specId: $specId) {
// 		id
// 	}
// }`,
// 		Variables: map[string]interface{}{
// 			"channelIds": channelIDs,
// 			"specId":     specID,
// 		},
// 	}

// 	if err := c.graphqlClient.ExecuteRequest(request, &response); err != nil {
// 		// return err
// 		return fmt.Errorf("PromoteCollector with YAML: %v", err)
// 	}

// 	if len(response.Errors) != 0 {
// 		return errors.New(response.Errors[0].Message)
// 	}

// 	return nil
// }

// CreateCollector - input appID, name, yaml - return Name, Spec, Config
func (c *HTTPClient) CreateCollector(appID string, yaml string) (*v1.AppCollectorInfo, error) {
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
	}
`,
		Variables: map[string]interface{}{
			"appId": appID,
			"spec":  yaml,
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
