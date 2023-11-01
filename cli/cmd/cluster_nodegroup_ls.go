package cmd

import (
	"reflect"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterNodeGroupList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List node groups",
		Long:  `List node groups`,
		RunE:  r.listNodeGroups,
		Args:  cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
	cmd.Flags().BoolVarP(&r.args.lsClusterNodeGroupsWatch, "watch", "w", false, "watch node groups")

	cmd.MarkFlagRequired("cluster-id")

	return cmd
}

func (r *runners) listNodeGroups(cmd *cobra.Command, args []string) error {
	nodeGroups, err := r.kotsAPI.ListClusterNodeGroups(args[0])
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list cluster node groups")
	}

	header := true
	if r.args.lsClusterNodeGroupsWatch {
		// Checks to see if the outputFormat is table
		if r.outputFormat != "table" {
			return errors.New("watch is only supported for table output")
		}

		nodeGroupsToPrint := make([]*types.ClusterNodeGroup, 0)

		// Prints the intial list of clusters
		if len(nodeGroups) == 0 {
			print.NoClusters(r.outputFormat, r.w)
		} else {
			nodeGroupsToPrint = append(nodeGroupsToPrint, nodeGroups...)
		}

		// Runs until ctrl C is recognized
		for range time.Tick(2 * time.Second) {
			newNodeGroups, err := r.kotsAPI.ListClusterNodeGroups(args[0])
			if err != nil {
				if err == promptui.ErrInterrupt {
					return errors.New("interrupted")
				}

				return errors.Wrap(err, "watch cluster node groups")
			}

			// Create a map from the IDs of the new clusters node groups
			newNodeGroupsMap := make(map[string]*types.ClusterNodeGroup)
			for _, newNodeGroup := range newNodeGroups {
				newNodeGroupsMap[newNodeGroup.ID] = newNodeGroup
			}

			// Create a map from the IDs of the old clusters
			oldNodeGroupsMap := make(map[string]*types.ClusterNodeGroup)
			for _, nodeGroup := range nodeGroups {
				oldNodeGroupsMap[nodeGroup.ID] = nodeGroup
			}

			// Check for new cluster node groups and print them
			for id, newNodeGroup := range newNodeGroupsMap {
				if oldNodeGroup, found := oldNodeGroupsMap[id]; !found {
					nodeGroupsToPrint = append(nodeGroupsToPrint, newNodeGroup)
				} else {
					// Check if properties of existing node group have changed
					if !reflect.DeepEqual(newNodeGroup, oldNodeGroup) {
						nodeGroupsToPrint = append(nodeGroupsToPrint, newNodeGroup)
					}
				}
			}

			// Check for removed cluster node groups and print them, changing their status to be "deleted"
			for id, nodeGroup := range oldNodeGroupsMap {
				if _, found := newNodeGroupsMap[id]; !found {
					nodeGroup.Status = types.ClusterNodeGroupStatusDeleted
					nodeGroupsToPrint = append(nodeGroupsToPrint, nodeGroup)
				}
			}

			// Prints the clusters
			if len(nodeGroupsToPrint) > 0 {
				print.ClusterNodeGroups(r.outputFormat, r.w, nodeGroupsToPrint, header)
				header = false // only print the header once
			}

			nodeGroups = newNodeGroups
			nodeGroupsToPrint = make([]*types.ClusterNodeGroup, 0)
		}
	}

	if len(nodeGroups) == 0 {
		return print.NoNodeGroups(r.outputFormat, r.w)
	}

	return print.ClusterNodeGroups(r.outputFormat, r.w, nodeGroups, true)
}
