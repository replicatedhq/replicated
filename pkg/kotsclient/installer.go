package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) ListInstallers(appID string) ([]types.InstallerSpec, error) {
	var response []types.InstallerSpec

	url := fmt.Sprintf("/v3/app/%s/installers", appID)
	err := c.DoJSON("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *VendorV3Client) CreateInstaller(appID string, yaml string) (*types.InstallerSpec, error) {
	request := types.CreateInstallerRequest{
		Yaml: yaml,
	}

	var response types.InstallerSpecResponse

	url := fmt.Sprintf("/v3/app/%s/installer", appID)
	err := c.DoJSON("POST", url, http.StatusCreated, request, &response)
	if err != nil {
		return nil, err
	}

	return &response.Body, nil
}

func (c *VendorV3Client) PromoteInstaller(appID string, sequence int64, channelID string, versionLabel string) error {
	request := types.PromoteInstallerRequest{
		Sequence:     sequence,
		VersionLabel: versionLabel,
		ChannelID:    channelID,
	}

	url := fmt.Sprintf("/v3/app/%s/installer/promote", appID)
	err := c.DoJSON("POST", url, http.StatusOK, request, nil)
	if err != nil {
		return err
	}

	return nil
}
