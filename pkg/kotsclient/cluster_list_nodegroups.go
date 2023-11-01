package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterNodeGroupsResponse struct {
	ClusterNodeGroups []*types.ClusterNodeGroup `json:"nodegroups"`
}

func (c VendorV3Client) ListClusterNodeGroups(clusterID string) ([]*types.ClusterNodeGroup, error) {
	resp := ListClusterNodeGroupsResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/%s/nodegroups", clusterID)

	err := c.DoJSON("GET", endpoint, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.ClusterNodeGroups, nil
}
