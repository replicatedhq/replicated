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
		Short: "Manage cluster add-ons.",
		Long: `The 'cluster addon' command allows you to manage add-ons installed on a test cluster. Add-ons are additional components or services that can be installed and configured to enhance or extend the functionality of the cluster.

You can use various subcommands to create, list, remove, or check the status of add-ons on a cluster. This command is useful for adding databases, object storage, monitoring, security, or other specialized tools to your cluster environment.`,
		Example: `# List all add-ons installed on a cluster
replicated cluster addon ls CLUSTER_ID_OR_NAME

# Remove an add-on from a cluster
replicated cluster addon rm CLUSTER_ID_OR_NAME --id ADDON_ID

# Create an object store bucket add-on for a cluster
replicated cluster addon create object-store CLUSTER_ID_OR_NAME --bucket-prefix mybucket

# List add-ons with JSON output
replicated cluster addon ls CLUSTER_ID_OR_NAME --output json`,
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
		}
		if addon.Status == types.ClusterAddonStatusError {
			return nil, errors.New("cluster addon failed to provision")
		}
		if time.Now().After(start.Add(duration)) {
			// In case of timeout, return the cluster and a WaitDurationExceeded error
			return addon, ErrWaitDurationExceeded
		}

		time.Sleep(time.Second * 5)
	}
}
