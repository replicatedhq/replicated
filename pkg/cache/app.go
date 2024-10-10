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

func (c *Cache) SetApp(app *types.App) error {
	c.Apps = append(c.Apps, *app)

	if err := c.Save(); err != nil {
		return errors.Wrap(err, "failed to save cache")
	}
	return nil
}
