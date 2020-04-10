package enterpriseclient

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

func (c HTTPClient) ListInstallers() ([]*enterprisetypes.Installer, error) {
	enterpriseInstallers := []*enterprisetypes.Installer{}
	err := c.doJSON("GET", "/v1/installers", 200, nil, &enterpriseInstallers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get installers")
	}

	return enterpriseInstallers, nil
}

func (c HTTPClient) CreateInstaller(yaml string) (*enterprisetypes.Installer, error) {
	type CreateInstallerRequest struct {
		Yaml string `json:"yaml"`
	}
	createInstallerRequest := CreateInstallerRequest{
		Yaml: yaml,
	}

	enterpriseInstaller := enterprisetypes.Installer{}
	err := c.doJSON("POST", "/v1/installer", 201, createInstallerRequest, &enterpriseInstaller)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create installer")
	}

	return &enterpriseInstaller, nil
}

func (c HTTPClient) UpdateInstaller(id string, yaml string) (*enterprisetypes.Installer, error) {
	type UpdateInstallerRequest struct {
		ID   string `json:"id"`
		Yaml string `json:"yaml"`
	}
	updateInstallerRequest := UpdateInstallerRequest{
		ID:   id,
		Yaml: yaml,
	}

	enterpriseInstaller := enterprisetypes.Installer{}

	err := c.doJSON("PUT", fmt.Sprintf("/v1/installer/%s", id), 200, updateInstallerRequest, &enterpriseInstaller)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update installer")
	}

	return &enterpriseInstaller, nil
}

func (c HTTPClient) RemoveInstaller(id string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/v1/installer/%s", id), 204, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete installer")
	}

	return nil
}

func (c HTTPClient) AssignInstaller(installerID string, channelID string) error {
	err := c.doJSON("POST", fmt.Sprintf("/v1/channelinstaller/%s/%s", installerID, channelID), 204, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to assign installer")
	}

	return nil
}
