package client

import (
	"errors"

	releases "github.com/replicatedhq/replicated/gen/go/v1"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListReleases(appID string, appType string) ([]types.ReleaseInfo, error) {
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
		shipReleases, err := c.ShipClient.ListReleases(appID)
		if err != nil {
			return nil, err
		}

		releaseInfos := make([]types.ReleaseInfo, 0, 0)
		for _, shipRelease := range shipReleases {
			activeChannels := make([]types.Channel, 0, 0)
			for _, shipReleaseChannel := range shipRelease.ActiveChannels {
				activeChannel := types.Channel{
					ID:          shipReleaseChannel.ID,
					Name:        shipReleaseChannel.Name,
					Description: shipReleaseChannel.Description,
				}

				activeChannels = append(activeChannels, activeChannel)
			}
			releaseInfo := types.ReleaseInfo{
				AppID:          shipRelease.AppID,
				CreatedAt:      shipRelease.CreatedAt,
				EditedAt:       shipRelease.EditedAt,
				Editable:       shipRelease.Editable,
				Sequence:       shipRelease.Sequence,
				Version:        shipRelease.Version,
				ActiveChannels: activeChannels,
			}

			releaseInfos = append(releaseInfos, releaseInfo)
		}

		return releaseInfos, nil
	} else if appType == "kots" {
		return c.KotsClient.ListReleases(appID)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) CreateRelease(appID string, appType string, yaml string) (*types.ReleaseInfo, error) {

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
	} else if appType == "kots" {
		return c.KotsClient.CreateRelease(appID, yaml)
	}

	return nil, errors.New("unknown app type")
}

func (c *Client) UpdateRelease(appID string, appType string, sequence int64, yaml string) error {

	if appType == "platform" {
		return c.PlatformClient.UpdateRelease(appID, sequence, yaml)
	} else if appType == "ship" {
		return c.ShipClient.UpdateRelease(appID, sequence, yaml)
	} else if appType == "kots" {
		return c.KotsClient.UpdateRelease(appID, sequence, yaml)
	}
	return errors.New("unknown app type")
}

func (c *Client) GetRelease(appID string, appType string, sequence int64) (*releases.AppRelease, error) {

	if appType == "platform" {
		return c.PlatformClient.GetRelease(appID, sequence)
	} else if appType == "ship" {
		return nil, errors.New("This feature is not supported for Ship applications.")
	} else if appType == "kots" {
		return c.KotsClient.GetRelease(appID, sequence)
	}
	return nil, errors.New("unknown app type")
}

func (c *Client) PromoteRelease(appID string, appType string, sequence int64, label string, notes string, required bool, channelIDs ...string) error {

	if appType == "platform" {
		return c.PlatformClient.PromoteRelease(appID, sequence, label, notes, required, channelIDs...)
	} else if appType == "ship" {
		return c.ShipClient.PromoteRelease(appID, sequence, label, notes, channelIDs...)
	} else if appType == "kots" {
		return c.KotsClient.PromoteRelease(appID, sequence, label, notes, required, channelIDs...)
	}
	return errors.New("unknown app type")
}

// data is a []byte describing a tarred yaml-dir, created by tarYAMLDir()
// this Client abstraction continue to spring more leaks :)
func (c *Client) LintRelease(appType string, data []byte) ([]types.LintMessage, error) {
	if appType == "platform" {
		return nil, errors.New("Linting is not yet supported in this CLI, please install github.com/replicatedhq/replicated-lint to lint this application")
	} else if appType == "ship" {
		return nil, errors.New("Linting is not supported for ship applications")
	} else if appType == "kots" {
		return c.KotsClient.LintRelease(data)
	}

	return nil, errors.New("unknown app type")
}
