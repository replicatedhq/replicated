package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type GetVMResponse struct {
	VM    *types.VM `json:"vm"`
	Error string    `json:"error"`
}

func (c *VendorV3Client) GetVM(id string) (*types.VM, error) {
	vm := GetVMResponse{}

	err := c.DoJSON(context.TODO(), "GET", fmt.Sprintf("/v3/vm/%s", id), http.StatusOK, nil, &vm)
	if err != nil {
		return nil, err
	}

	return vm.VM, nil
}
