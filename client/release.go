package client

import (
	"errors"

	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListReleases(appID string) ([]types.ReleaseInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformReleases, err := c.PlatformClient.ListReleases(appID)
		if err != nil {
			return nil, err
		}

		releaseInfos := make([]types.ReleaseInfo, 0, 0)
		for _, platformRelease := range platformReleases {
			activeChannels := make([]types.Channel, 0, 0)
			for _, platformReleaseChannel := range platformRelease.ActiveChannels {
				activeChannel := types.Channel{
					ID:          platformReleaseChannel.Id,
					Name:        platformReleaseChannel.Name,
					Description: platformReleaseChannel.Description,
				}

				activeChannels = append(activeChannels, activeChannel)
			}
			releaseInfo := types.ReleaseInfo{
				AppID:          platformRelease.AppId,
				CreatedAt:      platformRelease.CreatedAt,
				EditedAt:       platformRelease.EditedAt,
				Editable:       platformRelease.Editable,
				Sequence:       platformRelease.Sequence,
				Version:        platformRelease.Version,
				ActiveChannels: activeChannels,
			}

			releaseInfos = append(releaseInfos, releaseInfo)
		}

		return releaseInfos, nil
	} else if appType == "ship" {
		return c.ShipClient.ListReleases(appID)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) CreateRelease(appID string, yaml string) (*types.ReleaseInfo, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		platformReleaseInfo, err := c.PlatformClient.CreateRelease(appID, yaml)
		if err != nil {
			return nil, err
		}

		activeChannels := make([]types.Channel, 0, 0)

		for _, platformReleaseChannel := range platformReleaseInfo.ActiveChannels {
			activeChannel := types.Channel{
				ID:          platformReleaseChannel.Id,
				Name:        platformReleaseChannel.Name,
				Description: platformReleaseChannel.Description,
			}

			activeChannels = append(activeChannels, activeChannel)
		}
		return &types.ReleaseInfo{
			AppID:          platformReleaseInfo.AppId,
			CreatedAt:      platformReleaseInfo.CreatedAt,
			EditedAt:       platformReleaseInfo.EditedAt,
			Editable:       platformReleaseInfo.Editable,
			Sequence:       platformReleaseInfo.Sequence,
			Version:        platformReleaseInfo.Version,
			ActiveChannels: activeChannels,
		}, nil
	} else if appType == "ship" {
		return c.ShipClient.CreateRelease(appID, yaml)
	} else {
		return c.KotsClient.CreateRelease(appID, yaml)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateRelease(appID string, sequence int64, releaseOptions interface{}) error {
	return nil
}

func (c *Client) GetRelease(appID string, sequence int64) (interface{}, error) {
	return nil, nil
}

func (c *Client) PromoteRelease(appID string, sequence int64, label string, notes string, required bool, channelIDs ...string) error {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.PromoteRelease(appID, sequence, label, notes, required, channelIDs...)
	} else if appType == "ship" {
		return c.ShipClient.PromoteRelease(appID, sequence, label, notes, channelIDs...)
	}

	return errors.New("unknown app type")
}

func (c *Client) LintRelease(appID string, yaml string) ([]types.LintMessage, error) {
	appType, err := c.GetAppType(appID)
	if err != nil {
		return nil, err
	}

	if appType == "platform" {
		return nil, errors.New("Linting is not yet supported for Platform applications")
		// return c.PlatformClient.LintRelease(appID, yaml)
	} else if appType == "ship" {
		return c.ShipClient.LintRelease(appID, yaml)
	}

	return nil, errors.New("unknown app type")
}
