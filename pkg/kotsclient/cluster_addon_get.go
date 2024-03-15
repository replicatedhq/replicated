package kotsclient

import (
	"fmt"

	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GetClusterAddonResponse struct {
	Addon *types.ClusterAddon `json:"addon"`
}

func (c *VendorV3Client) GetClusterAddon(clusterID, id string) (*types.ClusterAddon, error) {
	resp := GetClusterAddonResponse{}

	addons, err := c.ListClusterAddons(clusterID)
	if err != nil {
		return nil, fmt.Errorf("list cluster addons: %v", err)
	}

	for _, addon := range addons {
		if addon.ID == id {
			resp.Addon = addon
			break
		}
	}

	if resp.Addon == nil {
		return nil, platformclient.ErrNotFound
	}

	return resp.Addon, nil
}
