package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

type UpdateVMRBACPolicyRequest struct {
	PolicyID string `json:"policy_id"`
}

func (c *VendorV3Client) UpdateVMRBACPolicy(vmID, policyID string) error {
	req := UpdateVMRBACPolicyRequest{
		PolicyID: policyID,
	}
	endpoint := fmt.Sprintf("/v3/vm/%s/rbac-policy", vmID)
	return c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusNoContent, req, nil)
}
