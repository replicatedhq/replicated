package client

import (
	"fmt"
	"net/http"

	apps "github.com/replicatedhq/replicated/gen/go/v1"
)

// ListApps returns all apps and their channels.
func (c *HTTPClient) ListApps() ([]apps.AppAndChannels, error) {
	appsAndChannels := make([]apps.AppAndChannels, 0)
	err := c.doJSON("GET", "/v1/apps", http.StatusOK, nil, &appsAndChannels)
	if err != nil {
		return nil, err
	}
	return appsAndChannels, nil
}

// GetApp resolves an app by either slug or ID.
func (c *HTTPClient) GetApp(slugOrID string) (*apps.App, error) {
	appsAndChannels, err := c.ListApps()
	if err != nil {
		return nil, fmt.Errorf("GetApp: %v", err)
	}
	for _, ac := range appsAndChannels {
		if ac.App.Slug == slugOrID || ac.App.Id == slugOrID {
			return ac.App, nil
		}
	}
	return nil, ErrNotFound
}

// CreateApp creates a new app with the given name and returns it.
func (c *HTTPClient) CreateApp(opts *AppOptions) (*apps.App, error) {
	reqBody := &apps.Body{Name: opts.Name}
	app := &apps.App{}
	err := c.doJSON("POST", "/v1/app", http.StatusCreated, reqBody, app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// DeleteApp deletes an app by id.
func (c *HTTPClient) DeleteApp(id string) error {
	endpoint := fmt.Sprintf("%s/v1/app/%s", c.apiOrigin, id)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteApp (%s %s): %v", req.Method, endpoint, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("DeleteApp (%s %s): status %d", req.Method, endpoint, resp.StatusCode)
	}
	return nil
}
