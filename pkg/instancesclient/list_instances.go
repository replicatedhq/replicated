package instancesclient

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
)

func (c InstancesClient) ListInstances(customer types.Customer) ([]types.CustomerInstance, error) {
	var instancesRaw []types.CustomerInstanceRaw
	err := c.HTTPClient.DoJSON(
		"GET",
		fmt.Sprintf("/v1/instances?customerSelector=%s&selectorType=id", customer.ID),
		200,
		nil,
		&instancesRaw,
	)

	if err != nil {
		return nil, errors.Wrap(err, "get instances from api")
	}

	var instances []types.CustomerInstance
	for _, instance := range instancesRaw {
		instances = append(instances, instance.WithInsights(customer.CreatedAt.Time))
	}

	return instances, nil
}
