package client

import (
	"time"

	"github.com/pkg/errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListCustomers(appID string, appType string) ([]types.Customer, error) {

	if appType == "platform" {
		return nil, errors.New("listing customers is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("listing customers is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.ListCustomers(appID)
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}

func (c *Client) CreateCustomer(appID, appType string, name string, channelID string, expiresIn time.Duration, isAirgapEnabled bool, isGitopsSupported bool, isSnapshotSupported bool, licenseType string) (*types.Customer, error) {
	if appType == "platform" {
		return nil, errors.New("creating customers is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("creating customers is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.CreateCustomer(name, appID, channelID, expiresIn, isAirgapEnabled, isGitopsSupported, isSnapshotSupported, licenseType)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) GetCustomerByName(appType string, appID, name string) (*types.Customer, error) {
	if appType == "platform" {
		return nil, errors.New("listing customers is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("listing customers is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.GetCustomerByName(appID, name)
	}

	return nil, errors.Errorf("unknown app type %q", appType)

}

func (c *Client) DownloadLicense(appType string, appID string, customerID string) ([]byte, error) {
	if appType == "platform" {
		return nil, errors.New("downloading customer licenses is not supported for platform applications")
	} else if appType == "ship" {
		return nil, errors.New("downloading  customer licenses is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.DownloadLicense(appID, customerID)
	}
	return nil, errors.Errorf("unknown app type %q", appType)
}
