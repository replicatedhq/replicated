package kotsclient

import (
	"context"
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
	err := c.DoJSON(context.TODO(), "DELETE", fmt.Sprintf("/v3/cluster/%s/port/%d?protocols=%s", clusterID, portNumber, urlProtocols), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Ports, nil
}
