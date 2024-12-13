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

func (r *runners) InitNetworkList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List test networks",
		Long:    ``,
		RunE:    r.listNetworks,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.lsNetworkStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsNetworkEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	cmd.Flags().BoolVarP(&r.args.lsNetworkWatch, "watch", "w", false, "watch networks")

	return cmd
}

func (r *runners) listNetworks(_ *cobra.Command, args []string) error {
	const longForm = "2006-01-02T15:04:05Z"
	var startTime, endTime *time.Time
	if r.args.lsNetworkStartTime != "" {
		st, err := time.Parse(longForm, r.args.lsNetworkStartTime)
		if err != nil {
			return errors.Wrap(err, "parse start time")
		}
		startTime = &st
	}
	if r.args.lsNetworkEndTime != "" {
		et, err := time.Parse(longForm, r.args.lsNetworkEndTime)
		if err != nil {
			return errors.Wrap(err, "parse end time")
		}
		endTime = &et
	}

	networks, err := r.kotsAPI.ListNetworks(startTime, endTime)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list networks")
	}

	header := true
	if r.args.lsNetworkWatch {

		// Checks to see if the outputFormat is table
		if r.outputFormat != "table" && r.outputFormat != "wide" {
			return errors.New("watch is only supported for table output")
		}

		networksToPrint := make([]*types.Network, 0)

		// Prints the intial list of networks
		if len(networks) == 0 {
			print.NoNetworks(r.outputFormat, r.w)
		} else {
			networksToPrint = append(networksToPrint, networks...)
		}

		// Runs until ctrl C is recognized
		for range time.Tick(2 * time.Second) {
			newNetworks, err := r.kotsAPI.ListNetworks(startTime, endTime)

			if err != nil {
				if err == promptui.ErrInterrupt {
					return errors.New("interrupted")
				}

				return errors.Wrap(err, "watch networks")
			}

			// Create a map from the IDs of the new networks
			newNetworkMap := make(map[string]*types.Network)
			for _, newNetwork := range newNetworks {
				newNetworkMap[newNetwork.ID] = newNetwork
			}

			// Create a map from the IDs of the old networks
			oldNetworkMap := make(map[string]*types.Network)
			for _, network := range networks {
				oldNetworkMap[network.ID] = network
			}

			// Check for new networks and print them
			for id, newNetwork := range newNetworkMap {
				if oldNetwork, found := oldNetworkMap[id]; !found {
					networksToPrint = append(networksToPrint, newNetwork)
				} else {
					// Check if properties of existing networks have changed
					if !reflect.DeepEqual(newNetwork, oldNetwork) {
						networksToPrint = append(networksToPrint, newNetwork)
					}
				}
			}

			// Check for removed networks and print them, changing their status to be "deleted"
			for id, network := range oldNetworkMap {
				if _, found := newNetworkMap[id]; !found {
					network.Status = types.StatusDeleted
					networksToPrint = append(networksToPrint, network)
				}
			}

			// Prints the netwworks
			if len(networksToPrint) > 0 {
				print.Networks(r.outputFormat, r.w, networksToPrint, header)
				header = false // only print the header once
			}

			networks = newNetworks
			networksToPrint = make([]*types.Network, 0)
		}
	}

	if len(networks) == 0 {
		return print.NoNetworks(r.outputFormat, r.w)
	}

	return print.Networks(r.outputFormat, r.w, networks, true)
}
