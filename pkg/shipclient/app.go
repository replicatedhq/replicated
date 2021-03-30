package shipclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListApps struct {
	Data   *ShipData          `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type ShipData struct {
	Ship *ShipAppsData `json:"ship"`
}

type ShipAppsData struct {
	ShipApps []*ShipApp `json:"apps"`
}

type ShipApp struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Slug     string        `json:"slug"`
	Channels []ShipChannel `json:"channel"`
}

const listAppsQuery = `
query {
  ship {
    apps {
      id
      name
      icon
      created
      updated
      isDefault
      isArchived
      slug
      channels {
	id,
	name,
	description
      }
    }
  }
}`

func (c *GraphQLClient) ListApps() ([]types.AppAndChannels, error) {
	response := GraphQLResponseListApps{}

	request := graphql.Request{
		Query: listAppsQuery,

		Variables: map[string]interface{}{},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	appsAndChannels := make([]types.AppAndChannels, 0, 0)
	for _, app := range response.Data.Ship.ShipApps {
		channels := make([]types.Channel, 0, 0)
		for _, shipChannel := range app.Channels {
			channel := types.Channel{
				ID:          shipChannel.ID,
				Name:        shipChannel.Name,
				Description: shipChannel.Description,
			}
			channels = append(channels, channel)
		}

		appAndChannels := types.AppAndChannels{
			App: &types.App{
				ID:   app.ID,
				Name: app.Name,
				Slug: app.Slug,
				Scheduler: "ship",
			},
			Channels: channels,
		}

		appsAndChannels = append(appsAndChannels, appAndChannels)
	}

	return appsAndChannels, nil
}

func (c *GraphQLClient) GetApp(appID string) (*types.App, error) {
	apps, err := c.ListApps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.App.ID == appID || app.App.Slug == appID {
			return app.App, nil
		}
	}

	return nil, errors.New("App not found")
}
