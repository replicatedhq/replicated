package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type updateModelsInCollectionRequest struct {
	AddModels    []string `json:"add_models"`
	RemoveModels []string `json:"remove_models"`
}

func (c *VendorV3Client) UpdateModelsInCollection(collectionID string, modelsToAdd []string, modelsToRemove []string) error {
	var reqBody = updateModelsInCollectionRequest{
		AddModels:    modelsToAdd,
		RemoveModels: modelsToRemove,
	}

	err := c.DoJSON("PATCH", fmt.Sprintf("/v3/models/collection/%s/models", collectionID), http.StatusOK, reqBody, nil)
	if err != nil {
		return errors.Wrap(err, "update models in collection")
	}

	return nil
}
