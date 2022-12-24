package instancesclient

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"strings"
	"time"
)

const threeDaysAgo = -3 * 24 * time.Hour

func (c InstancesClient) GetInstanceByIDPrefix(customer types.Customer, idPrefix string) (*types.CustomerInstance, error) {
	allInstances, err := c.ListInstances(customer)
	if err != nil {
		return nil, errors.Wrap(err, "list instances")
	}

	matchingInstances := make([]types.CustomerInstance, 0)
	for _, instance := range allInstances {
		if strings.HasPrefix(instance.Instance.ID, idPrefix) {
			matchingInstances = append(matchingInstances, instance)
		}
	}

	if len(matchingInstances) == 0 {
		return nil, errors.Errorf("no instance found with ID %q for customer %q", idPrefix, customer.Name)
	}

	if len(matchingInstances) > 1 {
		return nil, fmt.Errorf("%d instances for customer %q have prefix %q, please use full ID",
			len(matchingInstances), customer.Name, idPrefix)
	}
	return &matchingInstances[0], nil
}

func (c InstancesClient) GetInstanceUptime(instance types.CustomerInstance, startTime time.Time, uptimeInterval time.Duration) (*types.CustomerInstanceUptime, error) {
	uptime := types.CustomerInstanceUptime{}
	timeParam := startTime.Format(time.RFC3339)
	urlPath := fmt.Sprintf("/v1/instance/%s/appstatus?startTime=%s&interval=%.0fh", instance.Instance.ID, timeParam, uptimeInterval.Hours())

	err := c.HTTPClient.DoJSON(
		"GET",
		urlPath,
		200,
		nil,
		&uptime,
	)

	if err != nil {
		return nil, errors.Wrap(err, "get uptime")
	}
	return &uptime, nil
}
