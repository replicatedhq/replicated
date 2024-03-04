package kotsclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/replicatedhq/replicated/pkg/types"
)

type RemoveClusterPortResponse struct {
	Ports []*types.ClusterPort `json:"ports"`
}

func (c *VendorV3Client) RemoveClusterPort(clusterID string, portNumber int, protocols []string) ([]*types.ClusterPort, error) {
	urlProtocols := strings.Join(protocols, ",")

	resp := RemoveClusterPortResponse{}
	err := c.DoJSON("PUT", fmt.Sprintf("/v3/cluster/%s/ports/%d?protocols=%s", clusterID, portNumber, urlProtocols), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Ports, nil
}
