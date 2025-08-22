package kotsclient

import (
	"context"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type AddKOTSRegistryRequest struct {
	Provider       string   `json:"provider"`
	Endpoint       string   `json:"endpoint"`
	Slug           string   `json:"slug"`
	AuthType       string   `json:"authType"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	AppIds         []string `json:"appIds,omitempty"`
	SkipValidation bool     `json:"skipValidation"`
}

type AddKOTSRegistryResponse struct {
	Registry *types.Registry `json:"external_registry"`
}

func (c *VendorV3Client) AddKOTSRegistry(reqBody AddKOTSRegistryRequest) (*types.Registry, error) {
	registry := AddKOTSRegistryResponse{}
	err := c.DoJSON(context.TODO(), "POST", "/v3/external_registry", http.StatusCreated, reqBody, &registry)
	if err != nil {
		return nil, err
	}
	return registry.Registry, nil
}
