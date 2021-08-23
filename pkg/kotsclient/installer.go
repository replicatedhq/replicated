package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/replicatedhq/replicated/pkg/util"
)

func (c *VendorV3Client) ListInstallers(appID string) ([]types.InstallerSpec, error) {
	var response types.ListInstallersResponse

	url := fmt.Sprintf("/v3/app/%s/installers", appID)
	err := c.DoJSON("GET", url, http.StatusOK, nil, &response)
	if err != nil {
		return nil, err
	}

	for i, installerSpec := range response.Body {
		createdAtTime, err := util.ParseTime(installerSpec.CreatedAtString)
		if err != nil {
			return nil, errors.Wrap(err, "parsing time string to CreatedAt time")
		}
		response.Body[i].CreatedAt = util.Time{createdAtTime}
	}

	return response.Body, nil
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
