package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListClusterAddonsResponse struct {
	Addons []*types.ClusterAddon `json:"addons"`
}

func (c *VendorV3Client) ListClusterAddons(clusterID string) ([]*types.ClusterAddon, error) {
	resp := ListClusterAddonsResponse{}

	endpoint := fmt.Sprintf("/v3/cluster/%s/addons", clusterID)
	err := c.DoJSON(context.TODO(), "GET", endpoint, http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Addons, nil
}
