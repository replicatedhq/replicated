package enterpriseclient

import (
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
