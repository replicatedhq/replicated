package kotsclient

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func (c *VendorV3Client) RemoveRegistry(endpoint string) error {
	err := c.DoJSON("DELETE", fmt.Sprintf(`/v3/external_registry/%s`, url.QueryEscape(endpoint)), http.StatusNoContent, nil, nil)
	if err != nil {
		return errors.Wrap(err, "remove registry")
	}

	return nil
}
