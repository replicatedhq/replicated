package client

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c *Client) SetInstanceTags(appID string, appType string, customerID string, instanceID string, tags []types.Tag) (*types.Instance, error) {
	if appType == "kots" {
		instance, err := c.KotsClient.SetIntanceTags(appID, customerID, instanceID, tags)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set instance tags")
		}
		return instance, nil
	}

	return nil, errors.Errorf("unknown app type %q", appType)
}
