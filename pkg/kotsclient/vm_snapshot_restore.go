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

func (c *VendorV3Client) RestoreVMSnapshot(snapshotID string, ttl string) (*types.VM, error) {
	req := RestoreVMSnapshotRequest{TTL: ttl}

	var vm types.VM
	endpoint := fmt.Sprintf("/v3/snapshots/%s/restore", snapshotID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusCreated, req, &vm)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}
