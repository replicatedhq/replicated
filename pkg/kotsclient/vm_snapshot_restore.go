package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) RestoreVMSnapshot(vmID, snapshotID string) (*types.VM, error) {
	var vm types.VM
	endpoint := fmt.Sprintf("/v3/vm/%s/snapshot/%s", vmID, snapshotID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusCreated, nil, &vm)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}
