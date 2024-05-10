package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ExportClusterPortRequest struct {
	Port       int      `json:"port"`
	Protocols  []string `json:"protocols"`
	IsWildcard bool     `json:"is_wildcard"`
}

type ExposeClusterPortResponse struct {
	Port *types.ClusterPort `json:"port"`
}

func (c *VendorV3Client) ExposeClusterPort(clusterID string, portNumber int, protocols []string, isWildcard bool) (*types.ClusterPort, error) {
	req := ExportClusterPortRequest{
		Port:       portNumber,
		Protocols:  protocols,
		IsWildcard: isWildcard,
	}

	resp := ExposeClusterPortResponse{}
	err := c.DoJSON("POST", fmt.Sprintf("/v3/cluster/%s/port", clusterID), http.StatusCreated, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Port, nil
}
