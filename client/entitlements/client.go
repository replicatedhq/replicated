package entitlements

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type GraphQLClient struct {
	GQLServer *url.URL
	Token     string
	Logger    log.Logger
}

// GraphQLRequest is a json-serializable request to the graphql server
type GraphQLRequest struct {
	Query         string            `json:"query"`
	Variables     map[string]string `json:"variables"`
	OperationName string            `json:"operationName"`
}

// GraphQLError represents an error returned by the graphql server
type GraphQLError struct {
	Locations []map[string]interface{} `json:"locations"`
	Message   string                   `json:"message"`
	Code      string                   `json:"code"`
}

type GraphQLResponseCreateEntitlementSpec struct {
	Data   *CreateEntitlementSpecResponse `json:"data,omitempty"`
	Errors []GraphQLError                 `json:"errors,omitempty"`
}

type GraphQLResponseSetDefault struct {
	Data   map[string]interface{} `json:"data,omitempty"` // dont care
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type CreateEntitlementSpecResponse struct {
	CreateEntitlementSpec *EntitlementSpec `json:"createEntitlementSpec,omitempty"`
}

type EntitlementSpec struct {
	ID        string `json:"id,omitempty"`
	Spec      string `json:"spec,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type GraphQLResponseCreateEntitlementValue struct {
	Data   *CreateEntitlementValueResponse `json:"data,omitempty"`
	Errors []GraphQLError                  `json:"errors,omitempty"`
}

type CreateEntitlementValueResponse struct {
	CreateEntitlementValue *EntitlementValue `json:"createEntitlementValue,omitempty"`
}

type EntitlementValue struct {
	ID         string `json:"id,omitempty"`
	SpecID     string `json:"specId,omitempty"`
	CustomerID string `json:"customerId,omitempty"`
	Key        string `json:"key,omitempty"`
	Value      string `json:"value,omitempty"`
}

func (r GraphQLResponseCreateEntitlementSpec) GraphQLError() []GraphQLError {
	return r.Errors
}

func (r GraphQLResponseSetDefault) GraphQLError() []GraphQLError {
	return r.Errors
}

func (r GraphQLResponseCreateEntitlementValue) GraphQLError() []GraphQLError {
	return r.Errors
}

type Errer interface {
	GraphQLError() []GraphQLError
}

func (c *GraphQLClient) CreateEntitlementSpec(name string, spec string, appId string) (*EntitlementSpec, error) {
	requestObj := GraphQLRequest{
		Query: `
mutation($spec: String!, $name: String!, $appId: String!) {
  createEntitlementSpec(spec: $spec, name: $name, labels:[{key:"replicated.com/app", value:$appId}]) {
    id 
    spec 
    name 
    createdAt
  }
}`,
		Variables: map[string]string{"spec": spec, "name": name, "appId": appId},
	}
	response := GraphQLResponseCreateEntitlementSpec{}
	err := c.executeRequest(requestObj, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "execute request")
	}

	if err := c.checkErrors(response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementSpec, nil
}

func (c *GraphQLClient) SetDefaultEntitlementSpec(specID string) (map[string]interface{}, error) {
	requestObj := GraphQLRequest{
		Query: `
mutation($specId: ID!) {
  setDefaultEntitlementSpec(id: $specId)
}`,
		Variables: map[string]string{"specId": specID},
	}
	response := GraphQLResponseSetDefault{}
	err := c.executeRequest(requestObj, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "execute request")
	}

	if err := c.checkErrors(response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *GraphQLClient) SetEntitlementValue(customerID, specID, key, value, datatype, appId string) (*EntitlementValue, error) {
	requestObj := GraphQLRequest{
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
		Variables: map[string]string{
			"customerId": customerID,
			"specId":     specID,
			"key":        key,
			"value":      value,
			"type":       datatype,
			"appId":      appId,
		},
	}
	response := GraphQLResponseCreateEntitlementValue{}
	err := c.executeRequest(requestObj, &response)
	if err != nil {
		return nil, errors.Wrapf(err, "execute request")
	}

	if err := c.checkErrors(response); err != nil {
		return nil, err
	}

	return response.Data.CreateEntitlementValue, nil

}

func (c *GraphQLClient) executeRequest(
	requestObj GraphQLRequest,
	deserializeTarget interface{},
) error {
	debug := log.With(level.Debug(c.Logger), "type", "graphQLClient")
	body, err := json.Marshal(requestObj)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest("POST", c.GQLServer.String(), bodyReader)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	if resp == nil {
		return errors.New("nil response from gql")
	}
	if resp.Body == nil {
		return errors.New("nil response.Body from gql")
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	debug.Log("body", responseBody)
	if err != nil {
		return errors.Wrap(err, "marshal body")
	}
	if err := json.Unmarshal(responseBody, deserializeTarget); err != nil {
		return errors.Wrap(err, "unmarshal response")
	}

	return nil
}

func (c *GraphQLClient) checkErrors(errer Errer) error {
	if errer.GraphQLError() != nil && len(errer.GraphQLError()) > 0 {
		var multiErr *multierror.Error
		for _, err := range errer.GraphQLError() {
			multiErr = multierror.Append(multiErr, fmt.Errorf("%s: %s", err.Code, err.Message))

		}
		return multiErr.ErrorOrNil()
	}
	return nil
}
