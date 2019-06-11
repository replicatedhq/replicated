package shipclient

import (
	"time"

	"github.com/pkg/errors"
	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
)

type GraphQLResponseListCollectors struct {
	Data *SupportBundleSpecsData `json:"data,omitempty"`
}

type SupportBundleSpecsData struct {
	SupportBundleSpecs []SupportBundleSpec `json:"supportBundleSpecs"`
}

type SupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []v1.AppChannel `json:"channels"`
	Config    string          `json:"spec,omitempty"`
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

func (c *GraphQLClient) UpdateCollector(appID string, appType string, specID, yaml string) (interface{}, error) {
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

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *GraphQLClient) ListCollectors(appID string, appType string) ([]v1.AppCollectorInfo, error) {
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

	if err := c.ExecuteRequest(request, &response); err != nil {
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
func (c *GraphQLClient) GetCollector(appID string, appType string, id string) (*v1.AppCollectorInfo, error) {
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

// PromoteCollector assigns collector to a specified channel.
func (c *GraphQLClient) PromoteCollector(appID string, appType string, specID string, channelIDs ...string) error {
	response := graphql.ResponseErrorOnly{}

	request := graphql.Request{
		Query: `
mutation  promoteTroubleshootSpec($channelIds: [String], $specId: ID!) {
	promoteTroubleshootSpec(channelIds: $channelIds, specId: $specId) {
		id
	}
}`,
		Variables: map[string]interface{}{
			"channelIds": channelIDs,
			"specId":     specID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	if len(response.Errors) != 0 {
		return errors.New(response.Errors[0].Message)
	}

	return nil
}

// CreateCollector - input appID, name, yaml - return Name, Spec, Config
func (c *GraphQLClient) CreateCollector(appID string, appType string, yaml string) (*v1.AppCollectorInfo, error) {
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

	if err := c.ExecuteRequest(request, &response); err != nil {
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
	if err := c.ExecuteRequest(request, &finalizeSpecCreate); err != nil {
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
