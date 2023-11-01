package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type RemoveClusterNodeGroupsResponse struct {
	ClusterNodeGroups []*types.ClusterNodeGroup `json:"nodegroups"`
	Error             string                    `json:"error"`
}

func (c *VendorV3Client) RemoveClusterNodeGroup(nodeGroup *types.ClusterNodeGroup) ([]*types.ClusterNodeGroup, error) {
	endpoint := fmt.Sprintf("/v3/cluster/%s/nodegroup/%s", nodeGroup.ClusterID, nodeGroup.ID)

	resp := RemoveClusterNodeGroupsResponse{}
	err := c.DoJSON("DELETE", endpoint, http.StatusOK, nil, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				return nil, errors.New(resp.Error)
			}
		}

		return nil, err
	}

	return resp.ClusterNodeGroups, nil
}
