package kotsclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/graphql"
	"github.com/replicatedhq/replicated/pkg/types"
)

type GraphQLResponseListApps struct {
	Data   *KotsData          `json:"data,omitempty"`
	Errors []graphql.GQLError `json:"errors,omitempty"`
}

type KotsData struct {
	Kots *KotsAppsData `json:"kots"`
}

type KotsAppsData struct {
	KotsApps []*KotsApp `json:"apps"`
}

type KotsAppChannelData struct {
	ID string `json:"id"`
}

type KotsApp struct {
	ID       string                `json:"id"`
	Name     string                `json:"name"`
	Slug     string                `json:"slug"`
	Channels []*KotsAppChannelData `json: "channels"`
}

func (c *GraphQLClient) ListApps() ([]types.AppAndChannels, error) {
	response := GraphQLResponseListApps{}

	request := graphql.Request{
		Query: `
		query kots {
			kots {
			  apps {
				id
				name
				created
				updated
				isDefault
				isArchived
				slug
				channels {
				  id
				}
				isKotsApp
			  }
			}
		  }
		`,

		Variables: map[string]interface{}{},
	}

	if err := c.ExecuteRequest(request, &response); err != nil {
		return nil, err
	}

	appsAndChannels := make([]types.AppAndChannels, 0, 0)
	for _, kotsapp := range response.Data.Kots.KotsApps {

		appAndChannels := types.AppAndChannels{
			App: &types.App{
				ID:   kotsapp.ID,
				Name: kotsapp.Name,
				Slug: kotsapp.Slug,
			},
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
