package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type GetClusterNodeGroupResponse struct {
	ClusterNodeGroup *types.ClusterNodeGroup `json:"nodegroup"`
}

func (c VendorV3Client) GetClusterNodeGroup(id string) (*types.ClusterNodeGroup, error) {
	resp := GetClusterNodeGroupResponse{}
	endpoint := fmt.Sprintf("/v3/cluster/nodegroup/%s", id)

	err := c.DoJSON("GET", endpoint, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.ClusterNodeGroup, nil
}
