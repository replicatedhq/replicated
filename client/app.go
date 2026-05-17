package client

import (
	"context"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) ListApps(excludeChannels bool) ([]types.AppAndChannels, error) {
	platformApps, err := c.PlatformClient.ListApps()
	if err != nil {
		return nil, err
	}

	kotsApps, err := c.KotsClient.ListApps(context.TODO(), excludeChannels)
	if err != nil {
		return nil, err
	}

	apps := make([]types.AppAndChannels, 0)
	for _, platformApp := range platformApps {
		channels := make([]types.Channel, 0)
		for _, platformChannel := range platformApp.Channels {
			channel := types.Channel{
				ID:          platformChannel.Id,
				Name:        platformChannel.Name,
				Description: platformChannel.Description,
			}

			channels = append(channels, channel)
		}

		app := types.AppAndChannels{
			App: &types.App{
				ID:        platformApp.App.Id,
				Name:      platformApp.App.Name,
				Scheduler: platformApp.App.Scheduler,
				Slug:      platformApp.App.Slug,
			},
			Channels: channels,
		}

		apps = append(apps, app)
	}

	apps = append(apps, kotsApps...)

	return apps, nil
}

func (c *Client) GetApp(appID string) (interface{}, error) {
	app, _, err := c.GetAppType(context.TODO(), appID, true)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (c *Client) CreateApp(appOptions interface{}) (interface{}, error) {
	switch opts := appOptions.(type) {
	case string:
		return c.KotsClient.CreateKOTSApp(context.TODO(), opts)
	case kotsclient.CreateKOTSAppRequest:
		return c.KotsClient.CreateKOTSApp(context.TODO(), opts.Name)
	case *kotsclient.CreateKOTSAppRequest:
		if opts == nil {
			return nil, errors.New("create app options cannot be nil")
		}
		return c.KotsClient.CreateKOTSApp(context.TODO(), opts.Name)
	case platformclient.AppOptions:
		return c.PlatformClient.CreateApp(&opts)
	case *platformclient.AppOptions:
		if opts == nil {
			return nil, errors.New("create app options cannot be nil")
		}
		return c.PlatformClient.CreateApp(opts)
	default:
		return nil, errors.Errorf("unsupported app options type %T", appOptions)
	}
}

func (c *Client) DeleteApp(appID string) error {
	app, appType, err := c.GetAppType(context.TODO(), appID, true)
	if err != nil {
		return err
	}

	if appType == "platform" {
		return c.PlatformClient.DeleteApp(app.ID)
	} else if appType == "kots" {
		return c.KotsClient.DeleteKOTSApp(context.TODO(), app.ID)
	}

	return errors.Errorf("unknown app type %q", appType)
}
