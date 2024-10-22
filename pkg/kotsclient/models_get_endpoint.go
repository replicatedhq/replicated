package kotsclient

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type modelsGetEndpointResponse struct {
	Endpoint string `json:"endpoint"`
}

func (c *VendorV3Client) GetModelsEndpoint() (string, error) {
	var response = modelsGetEndpointResponse{}

	err := c.DoJSON(context.TODO(), "GET", "/v3/models/endpoint", http.StatusOK, nil, &response)
	if err != nil {
		return "", errors.Wrap(err, "get endpoint")
	}

	return response.Endpoint, nil
}
