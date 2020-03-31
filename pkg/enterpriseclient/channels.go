package enterpriseclient

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/enterprisetypes"
)

func (c HTTPClient) ListChannels() ([]*enterprisetypes.Channel, error) {
	enterpriseChannels := []*enterprisetypes.Channel{}
	err := c.doJSON("GET", "/v1/channels", 200, nil, &enterpriseChannels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channels")
	}

	return enterpriseChannels, nil
}
