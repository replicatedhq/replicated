package shipclient

import (
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
