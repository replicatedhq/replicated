package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) CreateVMSnapshot(vmID string) (*types.VMSnapshot, error) {
	var snapshot types.VMSnapshot
	endpoint := fmt.Sprintf("/v3/vm/%s/snapshot", vmID)
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusCreated, nil, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
