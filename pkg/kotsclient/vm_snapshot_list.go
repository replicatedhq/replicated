package kotsclient

import (
	"context"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListVMSnapshotsResponse struct {
	Snapshots []*types.VMSnapshot `json:"snapshots"`
}

func (c *VendorV3Client) ListVMSnapshots() ([]*types.VMSnapshot, error) {
	resp := ListVMSnapshotsResponse{}
	err := c.DoJSON(context.TODO(), "GET", "/v3/snapshots", http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Snapshots, nil
}
