package kotsclient

import (
	"net/http"
)

type TestKOTSRegistryRequest struct {
	Endpoint string `json:"endpoint"`
	Image    string `json:"image"`
}

type TestKOTSRegistryResponse struct {
	Status int `json:"status"`
}

func (c *VendorV3Client) TestKOTSRegistry(hostname string, image string) (int, error) {
	reqBody := &TestKOTSRegistryRequest{
		Endpoint: hostname,
		Image:    image,
	}

	resp := TestKOTSRegistryResponse{}
	err := c.DoJSON("PUT", "/v3/external_registry/test", http.StatusOK, reqBody, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Status, nil
}
