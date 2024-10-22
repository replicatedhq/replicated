package kotsclient

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type modelsResponse struct {
	Models []types.Model `json:"models"`
}

func (c *VendorV3Client) ListModels() ([]types.Model, error) {
	var response = modelsResponse{}

	err := c.DoJSON(context.TODO(), "GET", "/v3/models", http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list models")
	}

	results := make([]types.Model, 0)
	results = append(results, response.Models...)

	return results, nil
}
