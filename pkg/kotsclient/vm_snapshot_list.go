package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListVMSnapshotsResponse struct {
	Snapshots []*types.VMSnapshot `json:"snapshots"`
}

func (c *VendorV3Client) ListVMSnapshots(vmID string) ([]*types.VMSnapshot, error) {
	resp := ListVMSnapshotsResponse{}
	endpoint := fmt.Sprintf("/v3/vm/%s/snapshots", vmID)
	err := c.DoJSON(context.TODO(), "GET", endpoint, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Snapshots, nil
}
