package cache

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c Cache) GetApp(appSlugOrID string) (*types.App, error) {
	for _, app := range c.Apps {
		if app.Slug == appSlugOrID || app.ID == appSlugOrID {
			return &app, nil
		}
	}

	// App not found
	return nil, nil
}

// GetAppByID returns a cached app only for an exact app ID match. Unlike app
// IDs, slugs cannot be resolved safely from this cache because it is shared
// across API origins. The same slug can refer to different app IDs in
// production, staging, and local development environments.
func (c Cache) GetAppByID(appID string) (*types.App, error) {
	for _, app := range c.Apps {
		if app.ID == appID {
			return &app, nil
		}
	}

	return nil, nil
}

func (c *Cache) SetApp(app *types.App) error {
	c.Apps = append(c.Apps, *app)

	if err := c.Save(); err != nil {
		return errors.Wrap(err, "failed to save cache")
	}
	return nil
}
