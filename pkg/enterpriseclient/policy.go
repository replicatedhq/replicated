package enterpriseclient

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

func (c HTTPClient) ListPolicies() ([]*enterprisetypes.Policy, error) {
	enterprisePolicies := []*enterprisetypes.Policy{}
	err := c.doJSON("GET", "/v1/policies", 200, nil, &enterprisePolicies)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get policies")
	}

	return enterprisePolicies, nil
}

func (c HTTPClient) CreatePolicy(name string, description string, policy string) (*enterprisetypes.Policy, error) {
	type CreatePolicyRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Policy      string `json:"policy"`
	}
	createPolicyRequest := CreatePolicyRequest{
		Name:        name,
		Description: description,
		Policy:      policy,
	}

	enterprisePolicy := enterprisetypes.Policy{}
	err := c.doJSON("POST", "/v1/policy", 201, createPolicyRequest, &enterprisePolicy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create policy")
	}

	return &enterprisePolicy, nil
}

func (c HTTPClient) RemovePolicy(id string) error {
	fmt.Printf("id = %s\n", id)
	err := c.doJSON("DELETE", fmt.Sprintf("/v1/policy/%s", id), 204, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete policy")
	}

	return nil
}
