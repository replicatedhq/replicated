package kotsclient

import (
	"fmt"
	"net/http"
)

func (c *VendorV3Client) DeleteCollection(collectionID string) error {
	endpoint := fmt.Sprintf("/v3/models/collection?collectionId=%s", collectionID)
	err := c.DoJSON("DELETE", endpoint, http.StatusOK, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
