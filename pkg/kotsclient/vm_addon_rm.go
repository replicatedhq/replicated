package kotsclient

import (
	"context"
	"fmt"
	"net/http"
)

func (c *VendorV3Client) DeleteVMAddon(vmID, addonID string) error {
	endpoint := fmt.Sprintf("/v3/vm/%s/addons/%s", vmID, addonID)
	err := c.DoJSON(context.TODO(), "DELETE", endpoint, http.StatusNoContent, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
