package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type UpdateVMSnapshotRequest struct {
	TTL string `json:"ttl"`
}

func (c *VendorV3Client) UpdateVMSnapshot(snapshotID string, ttl string) (*types.VMSnapshot, error) {
	req := UpdateVMSnapshotRequest{TTL: ttl}
	var snapshot types.VMSnapshot
	endpoint := fmt.Sprintf("/v3/snapshots/%s", snapshotID)
	err := c.DoJSON(context.TODO(), "PUT", endpoint, http.StatusOK, req, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
