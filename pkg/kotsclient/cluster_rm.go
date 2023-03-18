package kotsclient

import (
	"fmt"
	"net/http"
)

type RemoveClusterRequest struct {
	ID string `json:"id"`
}

type RemoveClusterResponse struct {
	Error string `json:"error"`
}

func (c *VendorV3Client) RemoveCluster(id string) error {
	resp := RemoveClusterResponse{}
	err := c.DoJSON("DELETE", fmt.Sprintf("/v3/cluster/%s", id), http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}
	return nil
}
