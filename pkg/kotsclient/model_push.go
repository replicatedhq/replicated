package kotsclient

import "net/http"

type pushModelRequest struct {
	Name string `json:"name"`
}

func (c *VendorV3Client) PushModel(name string, path string) error {
	reqBody := pushModelRequest{
		Name: name,
	}

	// ignore path

	err := c.DoJSON("POST", "/v3/models/push-delete-this-route", http.StatusCreated, reqBody, nil)
	if err != nil {
		return err
	}

	return nil
}
