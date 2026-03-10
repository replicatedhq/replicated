package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateVMSnapshotRequest struct {
	Name string `json:"name,omitempty"`
}

func (c *VendorV3Client) CreateVMSnapshot(vmID, name string) (*types.VMSnapshot, error) {
	var snapshot types.VMSnapshot
	endpoint := fmt.Sprintf("/v3/vm/%s/snapshot", vmID)
	var body interface{}
	if name != "" {
		body = CreateVMSnapshotRequest{Name: name}
	}
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusCreated, body, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
