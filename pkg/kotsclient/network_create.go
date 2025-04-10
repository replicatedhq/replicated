package kotsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateNetworkRequest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	TTL     string `json:"ttl"`
}

type CreateNetworkResponse struct {
	Network *types.Network `json:"network"`
	Errors  []string       `json:"errors"`
}

type CreateNetworkDryRunResponse struct {
	TTL   *string                 `json:"ttl"`
	Error CreateNetworkErrorError `json:"error"`
}

type CreateNetworkOpts struct {
	Name    string
	Version string
	TTL     string
	DryRun  bool
}

type CreateNetworkErrorResponse struct {
	Error CreateNetworkErrorError `json:"error"`
}

type CreateNetworkErrorError struct {
	Message string `json:"message"`
}

func (c *VendorV3Client) CreateNetwork(opts CreateNetworkOpts) (*types.Network, *CreateNetworkErrorError, error) {
	req := CreateNetworkRequest{
		Name:    opts.Name,
		Version: opts.Version,
		TTL:     opts.TTL,
	}

	if opts.DryRun {
		return c.doCreateNetworkDryRunRequest(req)
	}
	return c.doCreateNetworkRequest(req)
}

func (c *VendorV3Client) doCreateNetworkRequest(req CreateNetworkRequest) (*types.Network, *CreateNetworkErrorError, error) {
	resp := CreateNetworkResponse{}
	endpoint := "/v3/network"
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &CreateNetworkErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, veResp); jsonErr != nil {
					return nil, nil, fmt.Errorf("unmarshal validation error response: %w", err)
				}
				return nil, &veResp.Error, nil
			}
		}

		return nil, nil, err
	}

	return resp.Network, nil, nil
}

func (c *VendorV3Client) doCreateNetworkDryRunRequest(req CreateNetworkRequest) (*types.Network, *CreateNetworkErrorError, error) {
	resp := CreateNetworkDryRunResponse{}
	endpoint := "/v3/network?dry-run=true"
	err := c.DoJSON(context.TODO(), "POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, nil, err
	}

	if resp.Error.Message != "" {
		return nil, &resp.Error, nil
	}
	return &types.Network{
		TTL: *resp.TTL,
	}, nil, nil
}
