package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type GetClusterResponse struct {
	Cluster *types.Cluster `json:"cluster"`
	Error   string         `json:"error"`
}

func (c *VendorV3Client) GetCluster(id string) (*types.Cluster, error) {
	cluster := GetClusterResponse{}

	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/cluster/%s", id), http.StatusOK, nil, &cluster)
	if err != nil {
		return nil, err
	}

	return cluster.Cluster, nil
}
