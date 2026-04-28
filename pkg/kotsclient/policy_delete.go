package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

func (c *VendorV3Client) DeletePolicy(id string) error {
	endpoint := fmt.Sprintf("/v3/policy/%s", id)
	return c.DoJSON(context.TODO(), "DELETE", endpoint, http.StatusOK, nil, nil)
}
