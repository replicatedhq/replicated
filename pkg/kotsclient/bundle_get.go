package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GetBundleResponse struct {
	Bundle *types.Bundle `json:"bundle"`
}

func (c *VendorV3Client) GetBundle(id string) (*types.Bundle, error) {
	resp := GetBundleResponse{}
	err := c.DoJSON("GET", fmt.Sprintf("/v3/ai/bundle/%s", id), http.StatusOK, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "load bundle")
	}

	return resp.Bundle, nil
}
