package cmd

import (
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddon(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "Manage cluster add-ons",
	}
	parent.AddCommand(cmd)

	return cmd
}

func waitForAddon(kotsRestClient *kotsclient.VendorV3Client, clusterID, id string, duration time.Duration) (*types.ClusterAddon, error) {
	start := time.Now()
	for {
		addon, err := kotsRestClient.GetClusterAddon(clusterID, id)
		if err != nil {
			return nil, errors.Wrap(err, "get cluster addon")
		}

		if addon.Status == types.ClusterAddonStatusRunning {
			return addon, nil
		} else if addon.Status == types.ClusterAddonStatusError {
			return nil, errors.New("cluster addon failed to provision")
		} else {
			if time.Now().After(start.Add(duration)) {
				// In case of timeout, return the cluster and a WaitDurationExceeded error
				return addon, ErrWaitDurationExceeded
			}
		}

		time.Sleep(time.Second * 5)
	}
}
