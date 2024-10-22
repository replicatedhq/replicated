package kotsclient

import (
	"context"
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

	url := fmt.Sprintf("/v3/cluster/%s", id)
	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}
	return nil
}
