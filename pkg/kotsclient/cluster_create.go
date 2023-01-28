package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateClusterRequest struct {
	Name string `json:"name"`
}

type CreateClusterResponse struct {
	Cluster *types.Cluster `json:"cluster"`
}

func (c *VendorV3Client) CreateCluster(name string) (*types.Cluster, error) {
	reqBody := &CreateClusterRequest{Name: name}
	cluster := CreateClusterResponse{}
	err := c.DoJSON("POST", "/v3/cluster", http.StatusCreated, reqBody, &cluster)
	if err != nil {
		return nil, err
	}
	return cluster.Cluster, nil
}
