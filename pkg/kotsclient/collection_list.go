package kotsclient

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type collectionResponse struct {
	Collections []types.ModelCollection `json:"collections"`
}

func (c *VendorV3Client) ListCollections() ([]types.ModelCollection, error) {
	var response = collectionResponse{}

	err := c.DoJSON(context.TODO(), "GET", "/v3/models/collections", http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list collections")
	}

	results := make([]types.ModelCollection, 0)
	results = append(results, response.Collections...)

	return results, nil
}
