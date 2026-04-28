package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition"`
}

type UpdatePolicyResponse struct {
	Policy *types.Policy `json:"policy"`
}

func (c *VendorV3Client) UpdatePolicy(id, name, description, definition string) (*types.Policy, error) {
	req := UpdatePolicyRequest{
		Name:        name,
		Description: description,
		Definition:  definition,
	}
	resp := UpdatePolicyResponse{}
	endpoint := fmt.Sprintf("/v3/policy/%s", id)
	if err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, req, &resp); err != nil {
		return nil, err
	}
	return resp.Policy, nil
}
