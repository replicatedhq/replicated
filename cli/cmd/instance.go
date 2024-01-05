package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitInstanceCommand(parent *cobra.Command) *cobra.Command {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage instances",
		Long:  `The instance command allows vendors to display and tag customer instances.`,
	}
	parent.AddCommand(instanceCmd)

	return instanceCmd
}

func findInstanceByNameOrID(nameOrID string, instances []types.Instance) (types.Instance, error) {
	var instance types.Instance

	isFound := false
	for _, i := range instances {
		if i.InstanceID == nameOrID || i.Name() == nameOrID {
			if isFound {
				return types.Instance{}, errors.Errorf("multiple instances found with name or id %q", nameOrID)
			}
			instance = i
			isFound = true
		}
	}

	if !isFound {
		return types.Instance{}, errors.Errorf("instance %q not found", nameOrID)
	}

	return instance, nil
}

func findInstancesByTags(tags []types.Tag, instances []types.Instance) []types.Instance {
	result := []types.Instance{}

	for _, instance := range instances {
		for _, instanceTag := range instance.Tags {
			for _, tag := range tags {
				if instanceTag.Key == tag.Key && instanceTag.Value == tag.Value {
					result = append(result, instance)
					break
				}
			}
		}
	}

	return result
}
