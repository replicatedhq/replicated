package shipclient

import (
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseCreateEntitlementSpec struct {
	Data   *CreateEntitlementSpecResponse `json:"data,omitempty"`
	Errors []graphql.GQLError             `json:"errors,omitempty"`
}

type CreateEntitlementSpecResponse struct {
	CreateEntitlementSpec *types.EntitlementSpec `json:"createEntitlementSpec,omitempty"`
}

type GraphQLResponseSetDefault struct {
	Data   map[string]interface{} `json:"data,omitempty"` // dont care
	Errors []graphql.GQLError     `json:"errors,omitempty"`
}

type GraphQLResponseCreateEntitlementValue struct {
	Data   *CreateEntitlementValueResponse `json:"data,omitempty"`
	Errors []graphql.GQLError              `json:"errors,omitempty"`
}

type CreateEntitlementValueResponse struct {
	CreateEntitlementValue *types.EntitlementValue `json:"createEntitlementValue,omitempty"`
}

const createEntitlementSpecQuery = `
mutation createEntitlementSpec($spec: String!, $name: String!, $appId: String!) {
  createEntitlementSpec(spec: $spec, name: $name, labels:[{key:"replicated.com/app", value:$appId}]) {
    id
    spec
    name
    createdAt
  }
}`

func (c *GraphQLClient) CreateEntitlementSpec(appID string, name string, spec string) (*types.EntitlementSpec, error) {
	response := GraphQLResponseCreateEntitlementSpec{}
	request := graphql.Request{
		Query: createEntitlementSpecQuery,
		Variables: map[string]interface{}{
			"spec":  spec,
			"name":  name,
			"appId": appID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementSpec, nil
}

const setDefaultEntitlementSpecQuery = `
mutation setDefaultEntitlementSpec($specId: ID!) {
  setDefaultEntitlementSpec(id: $specId)
}`

func (c *GraphQLClient) SetDefaultEntitlementSpec(specID string) error {
	response := GraphQLResponseSetDefault{}
	request := graphql.Request{
		Query: setDefaultEntitlementSpecQuery,
		Variables: map[string]interface{}{
			"specId": specID,
		},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return err
	}

	return nil
}

const setEntitlementValueQuery = `
mutation($customerId: String!, $specId: String!, $key: String!, $value: String!, $type: String!, $appId: String!) {
  createEntitlementValue(customerId: $customerId, specId: $specId, key: $key, value: $value, labels: [{key: "type", value: $type},{key:"replicated.com/app", value:$appId}]) {
    id
    key
    value
    labels {
      key
      value
    }
  }
}`

func (c *GraphQLClient) SetEntitlementValue(customerID string, specID string, key string, value string, datatype string, appId string) (*types.EntitlementValue, error) {
	response := GraphQLResponseCreateEntitlementValue{}
	request := graphql.Request{
		Query: setEntitlementValueQuery,
		Variables: map[string]interface{}{
			"customerId": customerID,
			"specId":     specID,
			"key":        key,
			"value":      value,
			"type":       datatype,
			"appId":      appId,
		},
	}
	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementValue, nil

}
