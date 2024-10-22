package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

type RemoveVMRequest struct {
	ID string `json:"id"`
}

type RemoveVMResponse struct {
	Error string `json:"error"`
}

func (c *VendorV3Client) RemoveVM(id string) error {
	resp := RemoveClusterResponse{}

	url := fmt.Sprintf("/v3/vm/%s", id)
	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		return err
	}
	return nil
}
