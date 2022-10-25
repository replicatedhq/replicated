package kotsclient

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type kotsRegistryResponse struct {
	Registries []types.Registry `json:"external_registries"`
}

func (c *VendorV3Client) ListRegistries() ([]types.Registry, error) {
	var response = kotsRegistryResponse{}

	err := c.DoJSON("GET", "/v3/external_registries", http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list registries")
	}

	results := make([]types.Registry, 0)
	results = append(results, response.Registries...)

	return results, nil
}
