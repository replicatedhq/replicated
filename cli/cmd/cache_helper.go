package cmd

import (
	"context"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
)

func getApp(appSlugOrID string, kotsClient *kotsclient.VendorV3Client) (*types.App, error) {
	app, err := cache.GetApp(appSlugOrID)
	if err == nil && app != nil {
		return app, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "get app from cache")
	}

	if app == nil {
		a, err := kotsClient.GetApp(context.TODO(), appSlugOrID, true)
		if err != nil {
			return nil, errors.Wrap(err, "get app from api")
		}

		if err := cache.SetApp(a); err != nil {
			return nil, errors.Wrap(err, "set app in cache")
		}

		return a, nil
	}

	return nil, errors.New("app not found")
}
