package shipclient

import (
	v1 "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListCollectors struct {
	Data *SupportBundleSpecsData `json:"data,omitempty"`
}

type SupportBundleSpecsData struct {
	SupportBundleSpecs []*SupportBundleSpec `json:"supportBundleSpecs"`
}

type SupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []types.Channel `json:"channels"`
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
	Name string `json:"name"`
}

type GraphQLResponseCreateCollector struct {
	Data *SupportBundleCreateSpecData `json:"data,omitempty"`
}

type SupportBundleCreateSpecData struct {
	CreateSupportBundleSpec *CreateSupportBundleSpec `json:"createSupportBundleSpec"`
}

type CreateSupportBundleSpec struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Config string `json:"spec,omitempty"`
}

// PLATFORM
type PlatformGQLResponseListCollectors struct {
	Data *PlatformSupportBundleSpecsData `json:"data,omitempty"`
}

type PlatformSupportBundleSpecsData struct {
	PlatformSupportBundleSpecs []*PlatformSupportBundleSpec `json:"supportBundleSpecs"`
}

type PlatformSupportBundleSpec struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreatedAt string          `json:"createdAt"`
	Channels  []types.Channel `json:"platformChannels"`
	Config    string          `json:"spec,omitempty"`
}

type PlatformChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetCollector returns a collector's properties.
func (c *GraphQLClient) GetCollector(appID string, id string) (*v1.AppCollectorInfo, error) {
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

	if err := c.ExecuteRequest(request, &response); err != nil {
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

// PromoteCollector assigns collector to a specified channel.
func (c *GraphQLClient) PromoteCollector(appID string, specID string, channelIDs ...string) error {
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

	return nil

}

// CreateCollector creates a new collector based on given yaml and name
func (c *GraphQLClient) CreateCollector(appID string, name string, yaml string) (*v1.AppCollectorInfo, error) {
	response := GraphQLResponseCreateCollector{}

	request := graphql.Request{
		Query: `
		mutation createSupportBundleSpec($name: String, $appId: String, $spec: String) {
		  createSupportBundleSpec(name: $name, appId: $appId, spec: $spec) {
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
			"name":  name,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	request = graphql.Request{
		Query: `
		mutation updateSupportBundleSpec($id: ID!, $spec: String!, $isArchived: Boolean) {
		  updateSupportBundleSpec(id: $id, spec: $spec, isArchived: $isArchived) {
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

func (c *GraphQLClient) UpdateCollector(appID string, specID, yaml string) (interface{}, error) {
	response := GraphQLResponseUpdateCollector{}

	request := graphql.Request{
		Query: `
mutation updateSupportBundleSpec($id: ID!, $spec: String!, $isArchived: Boolean) {
	updateSupportBundleSpec(id: $id, spec: $spec, isArchived: $isArchived) {
		id
		spec
		createdAt
		updatedAt
		isArchived
	}
}
`,

		Variables: map[string]interface{}{
			"id":   specID,
			"spec": yaml,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *GraphQLClient) UpdateCollectorName(appID string, specID, name string) (interface{}, error) {
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

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
