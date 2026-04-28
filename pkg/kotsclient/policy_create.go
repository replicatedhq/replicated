package kotsclient

import (
	"context"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition"`
}

type CreatePolicyResponse struct {
	Policy *types.Policy `json:"policy"`
}

func (c *VendorV3Client) CreatePolicy(name, description, definition string) (*types.Policy, error) {
	req := CreatePolicyRequest{
		Name:        name,
		Description: description,
		Definition:  definition,
	}
	resp := CreatePolicyResponse{}
	if err := c.DoJSON(context.TODO(), "POST", "/v3/policy", http.StatusOK, req, &resp); err != nil {
		return nil, err
	}
	return resp.Policy, nil
}
