package shipclient

import "github.com/replicatedhq/replicated/pkg/types"

type GraphQLResponseCreateEntitlementSpec struct {
	Data   *CreateEntitlementSpecResponse `json:"data,omitempty"`
	Errors []GraphQLError                 `json:"errors,omitempty"`
}

type CreateEntitlementSpecResponse struct {
	CreateEntitlementSpec *types.EntitlementSpec `json:"createEntitlementSpec,omitempty"`
}

type GraphQLResponseSetDefault struct {
	Data   map[string]interface{} `json:"data,omitempty"` // dont care
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type GraphQLResponseCreateEntitlementValue struct {
	Data   *CreateEntitlementValueResponse `json:"data,omitempty"`
	Errors []GraphQLError                  `json:"errors,omitempty"`
}

type CreateEntitlementValueResponse struct {
	CreateEntitlementValue *types.EntitlementValue `json:"createEntitlementValue,omitempty"`
}

func (c *GraphQLClient) CreateEntitlementSpec(appID string, name string, spec string) (*types.EntitlementSpec, error) {
	response := GraphQLResponseCreateEntitlementSpec{}
	request := GraphQLRequest{
		Query: `
mutation createEntitlementSpec($spec: String!, $name: String!, $appId: String!) {
  createEntitlementSpec(spec: $spec, name: $name, labels:[{key:"replicated.com/app", value:$appId}]) {
    id
    spec
    name
    createdAt
  }
}`,
		Variables: map[string]interface{}{
			"spec":  spec,
			"name":  name,
			"appId": appID,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementSpec, nil
}

func (c *GraphQLClient) SetDefaultEntitlementSpec(specID string) error {
	response := GraphQLResponseSetDefault{}
	request := GraphQLRequest{
		Query: `
mutation setDefaultEntitlementSpec($specId: ID!) {
  setDefaultEntitlementSpec(id: $specId)
}`,
		Variables: map[string]interface{}{
			"specId": specID,
		},
	}

	if err := c.executeRequest(request, &response); err != nil {
		return err
	}

	return nil
}

func (c *GraphQLClient) SetEntitlementValue(customerID string, specID string, key string, value string, datatype string, appId string) (*types.EntitlementValue, error) {
	response := GraphQLResponseCreateEntitlementValue{}
	request := GraphQLRequest{
		Query: `
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
}`,
		Variables: map[string]interface{}{
			"customerId": customerID,
			"specId":     specID,
			"key":        key,
			"value":      value,
			"type":       datatype,
			"appId":      appId,
		},
	}
	if err := c.executeRequest(request, &response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementValue, nil

}
