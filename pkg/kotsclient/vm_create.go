package kotsclient

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type CreateVMRequest struct {
	Name         string      `json:"name"`
	Distribution string      `json:"distribution"`
	Version      string      `json:"version"`
	Count        int         `json:"count"`
	DiskGiB      int64       `json:"disk_gib"`
	TTL          string      `json:"ttl"`
	InstanceType string      `json:"instance_type"`
	Tags         []types.Tag `json:"tags"`
}

type CreateVMResponse struct {
	VMs                    []*types.VM       `json:"vms"`
	Errors                 []string          `json:"errors"`
	SupportedDistributions map[string]string `json:"supported_distributions"`
}

type CreateVMDryRunResponse struct {
	TotalCost *int64             `json:"total_cost"`
	TTL       *string            `json:"ttl"`
	Error     CreateVMErrorError `json:"error"`
}

type CreateVMOpts struct {
	Name         string
	Distribution string
	Version      string
	Count        int
	DiskGiB      int64
	TTL          string
	InstanceType string
	Tags         []types.Tag
	DryRun       bool
}

type CreateVMErrorResponse struct {
	Error CreateVMErrorError `json:"error"`
}

type CreateVMErrorError struct {
	Message         string             `json:"message"`
	MaxDiskGiB      int64              `json:"maxDiskGiB,omitempty"`
	ValidationError *VMValidationError `json:"validationError,omitempty"`
}

type VMValidationError struct {
	Errors                 []string           `json:"errors"`
	SupportedDistributions []*types.VMVersion `json:"supported_distributions"`
}

func (c *VendorV3Client) CreateVM(opts CreateVMOpts) ([]*types.VM, *CreateVMErrorError, error) {
	req := CreateVMRequest{
		Name:         opts.Name,
		Distribution: opts.Distribution,
		Version:      opts.Version,
		Count:        opts.Count,
		DiskGiB:      opts.DiskGiB,
		TTL:          opts.TTL,
		InstanceType: opts.InstanceType,
		Tags:         opts.Tags,
	}

	if opts.DryRun {
		return c.doCreateVMDryRunRequest(req)
	}
	return c.doCreateVMRequest(req)
}

func (c *VendorV3Client) doCreateVMRequest(req CreateVMRequest) ([]*types.VM, *CreateVMErrorError, error) {
	resp := CreateVMResponse{}
	endpoint := "/v3/vm"
	err := c.DoJSON("POST", endpoint, http.StatusCreated, req, &resp)
	if err != nil {
		// if err is APIError and the status code is 400, then we have a validation error
		// and we can return the validation error
		if apiErr, ok := errors.Cause(err).(platformclient.APIError); ok {
			if apiErr.StatusCode == http.StatusBadRequest {
				veResp := &CreateVMErrorResponse{}
				if jsonErr := json.Unmarshal(apiErr.Body, veResp); jsonErr != nil {
					return nil, nil, fmt.Errorf("unmarshal validation error response: %w", err)
				}
				return nil, &veResp.Error, nil
			}
		}

		return nil, nil, err
	}

	return resp.VMs, nil, nil
}

func (c *VendorV3Client) doCreateVMDryRunRequest(req CreateVMRequest) ([]*types.VM, *CreateVMErrorError, error) {
	resp := CreateVMDryRunResponse{}
	endpoint := "/v3/vm?dry-run=true"
	err := c.DoJSON("POST", endpoint, http.StatusOK, req, &resp)
	if err != nil {
		return nil, nil, err
	}

	if resp.Error.Message != "" {
		return nil, &resp.Error, nil
	}
	vms := []*types.VM{
		{
			EstimatedCost: *resp.TotalCost,
			TTL:           *resp.TTL,
		},
	}

	return vms, nil, nil
}
