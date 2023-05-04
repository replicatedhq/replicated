package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"
)

func (c *VendorV3Client) ArchiveCustomer(customerID string) error {
	err := c.DoJSON("POST", fmt.Sprintf(`/v3/customer/%s/archive`, url.QueryEscape(customerID)), http.StatusNoContent, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
