package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ExportVMPortRequest struct {
	Port      int      `json:"port"`
	Protocols []string `json:"protocols"`
}

type ExposeVMPortResponse struct {
	Port *types.VMPort `json:"port"`
}

func (c *VendorV3Client) ExposeVMPort(vmID string, portNumber int, protocols []string) (*types.VMPort, error) {
	req := ExportVMPortRequest{
		Port:      portNumber,
		Protocols: protocols,
	}

	resp := ExposeVMPortResponse{}
	err := c.DoJSON(context.TODO(), "POST", fmt.Sprintf("/v3/vm/%s/port", vmID), http.StatusCreated, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Port, nil
}
