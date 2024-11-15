package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) CreateInstaller(appId string, appType string, yaml string) (*types.InstallerSpec, error) {
	if appType == "platform" {
		return nil, errors.Errorf("Kubernetes Installers are not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.CreateInstaller(appId, yaml)
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) ListInstallers(appId string, appType string) ([]types.InstallerSpec, error) {

	if appType == "platform" {
		return nil, errors.Errorf("Kubernetes Installers are not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.ListInstallers(appId)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) PromoteInstaller(appId string, appType string, sequence int64, channelID string, versionLabel string) error {
	if appType == "platform" {
		return errors.Errorf("Kubernetes Installers are not supported for platform applications")
	} else if appType == "kots" {
		return c.KotsClient.PromoteInstaller(appId, sequence, channelID, versionLabel)
	}

	return errors.Errorf("unknown app type %q", appType)

}
