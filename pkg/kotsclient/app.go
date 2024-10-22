package kotsclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	kotsclienttypes "github.com/replicatedhq/replicated/pkg/kotsclient/types"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *VendorV3Client) ListApps(ctx context.Context, excludeChannels bool) ([]types.AppAndChannels, error) {
	var response = kotsclienttypes.KotsAppResponse{}

	err := c.DoJSON(ctx, "GET", fmt.Sprintf("/v3/apps?excludeChannels=%t", excludeChannels), http.StatusOK, nil, &response)
	if err != nil {
		return nil, errors.Wrap(err, "list apps")
	}

	results := make([]types.AppAndChannels, 0)
	for _, kotsApp := range response.Apps {
		app := types.AppAndChannels{
			App: &types.App{
				ID:           kotsApp.Id,
				Name:         kotsApp.Name,
				Slug:         kotsApp.Slug,
				IsFoundation: kotsApp.IsFoundation,
				Scheduler:    "kots",
			},
			Channels: kotsApp.Channels,
		}
		results = append(results, app)
	}

	return results, nil
}

func (c *VendorV3Client) GetApp(appID string, excludeChannels bool) (*types.App, error) {
	apps, err := c.ListApps(context.TODO(), excludeChannels)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.App.ID == appID || app.App.Slug == appID {
			return &types.App{
				ID:           app.App.ID,
				Name:         app.App.Name,
				Slug:         app.App.Slug,
				IsFoundation: app.App.IsFoundation,
				Scheduler:    "kots",
			}, nil
		}
	}

	return nil, errors.New("App not found: " + appID)
}
