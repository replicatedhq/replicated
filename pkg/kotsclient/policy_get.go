package kotsclient

import (
	"fmt"

	"github.com/replicatedhq/replicated/pkg/types"
)

// GetPolicyByNameOrID returns the policy matching the given name or ID.
func (c *VendorV3Client) GetPolicyByNameOrID(nameOrID string) (*types.Policy, error) {
	policies, err := c.ListPolicies()
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	for _, p := range policies {
		if p.ID == nameOrID || p.Name == nameOrID {
			return p, nil
		}
	}

	return nil, fmt.Errorf("policy %q not found", nameOrID)
}
