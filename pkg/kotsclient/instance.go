package kotsclient

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

type InstanceTags struct {
	Tags []types.Tag `json:"tags"`
}

func (c *VendorV3Client) SetIntanceTags(appID string, customerID string, instanceID string, tags []types.Tag) (*types.Instance, error) {
	resp := struct {
		CustomerInstance types.Instance `json:"customerInstance"`
	}{}

	path := fmt.Sprintf("/v3/app/%s/customer/%s/instance/%s/tags", appID, customerID, instanceID)
	payload := InstanceTags{
		Tags: tags,
	}

	err := c.DoJSON("PUT", path, http.StatusOK, &payload, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set instance tags")
	}

	return &resp.CustomerInstance, nil
}
