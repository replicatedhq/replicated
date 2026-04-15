package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListPoliciesResponse struct {
	Policies []*types.Policy `json:"policies"`
}

// ListPolicies returns all RBAC policies for the authenticated team.
func (c *VendorV3Client) ListPolicies() ([]*types.Policy, error) {
	resp := ListPoliciesResponse{}
	if err := c.DoJSON(context.TODO(), "GET", "/v3/policies", http.StatusOK, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Policies, nil
}

// GetPolicyByName returns the RBAC policy with the given name, or an error if
// no policy with that name exists for the team.
func (c *VendorV3Client) GetPolicyByName(name string) (*types.Policy, error) {
	policies, err := c.ListPolicies()
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	for _, p := range policies {
		if p.Name == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("policy %q not found", name)
}
