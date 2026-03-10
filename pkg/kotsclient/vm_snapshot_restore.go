package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type RestoreVMSnapshotRequest struct {
	PublicKeys []string `json:"public_keys,omitempty"`
	TTL        string   `json:"ttl,omitempty"`
}

func (c *VendorV3Client) RestoreVMSnapshot(vmID, snapshotID string, publicKeys []string, ttl string) (*types.VM, error) {
	var req *RestoreVMSnapshotRequest
	if len(publicKeys) > 0 {
		req = &RestoreVMSnapshotRequest{PublicKeys: publicKeys}
	}

	if ttl != "" {
		req.TTL = ttl
	}

	var vm types.VM
	endpoint := fmt.Sprintf("/v3/vm/%s/snapshot/%s", vmID, snapshotID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusCreated, req, &vm)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}
