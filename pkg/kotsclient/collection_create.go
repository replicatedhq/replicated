package kotsclient

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type collectionCreateResponse struct {
	Collection *types.ModelCollection `json:"collection"`
}

type collectionCreeateRequest struct {
	Name string `json:"name"`
}

func (c *VendorV3Client) CreateCollection(name string) (*types.ModelCollection, error) {
	var response = collectionCreateResponse{}

	reqBody := collectionCreeateRequest{
		Name: name,
	}

	err := c.DoJSON("POST", "/v3/models/collection", http.StatusCreated, reqBody, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list collections")
	}

	return response.Collection, nil
}
