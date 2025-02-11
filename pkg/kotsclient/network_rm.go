package kotsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
)

type RemoveNetworkRequest struct {
	ID string `json:"id"`
}

type RemoveNetworkResponse struct {
	Error string `json:"error"`
}

type RemoveNetworkErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	}
}

func (c *VendorV3Client) RemoveNetwork(id string) error {
	resp := RemoveNetworkResponse{}

	url := fmt.Sprintf("/v3/network/%s", id)
	err := c.DoJSON(context.TODO(), "DELETE", url, http.StatusOK, nil, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				errResp := &RemoveNetworkErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, errResp); jsonErr != nil {
					return fmt.Errorf("unmarshal error response: %w", err)
				}
				return errors.New(errResp.Error.Message)
			}
		}
		return err
	}
	return nil
}
