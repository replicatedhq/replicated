package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterPortsResponse struct {
	Ports []*types.ClusterPort `json:"ports"`
}

func (c *VendorV3Client) ListClusterPorts(clusterID string) ([]*types.ClusterPort, error) {
	resp := ListClusterPortsResponse{}
	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/cluster/%s/ports", clusterID), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Ports, nil
}
