package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) RemoveModel(modelName string) ([]types.Model, error) {
	var response = modelsResponse{}

	err := c.DoJSON(context.TODO(), "DELETE", fmt.Sprintf("/v3/models/%s", modelName), http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list models")
	}

	results := make([]types.Model, 0)
	results = append(results, response.Models...)

	return results, nil
}
