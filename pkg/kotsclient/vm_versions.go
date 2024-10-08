package kotsclient

import (
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

type ListVMVersionsResponse struct {
	Versions []*types.VMVersion `json:"cluster-versions"`
}

func (c *VendorV3Client) ListVMVersions() ([]*types.VMVersion, error) {
	versions := ListVMVersionsResponse{}
	err := c.DoJSON("GET", "/v3/vm/versions", http.StatusOK, nil, &versions)
	if err != nil {
		return nil, err
	}
	return versions.Versions, nil
}
