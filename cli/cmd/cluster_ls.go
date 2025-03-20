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

func (r *runners) InitClusterList(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List test clusters.",
		Long: `The 'cluster ls' command lists all test clusters. This command provides information about the clusters, such as their status, name, distribution, version, and creation time. The output can be formatted in different ways, depending on your needs.

You can filter the list of clusters by time range and status (e.g., show only terminated clusters). You can also watch clusters in real-time, which updates the list every few seconds.

Clusters that have been deleted will be shown with a 'deleted' status.`,
		Example: `# List all clusters with default table output
replicated cluster ls

# Show clusters created after a specific date
replicated cluster ls --start-time 2023-01-01T00:00:00Z

# Watch for real-time updates
replicated cluster ls --watch

# List clusters with JSON output
replicated cluster ls --output json

# List only terminated clusters
replicated cluster ls --show-terminated

# List clusters with wide table output
replicated cluster ls --output wide`,
		RunE: r.listClusters,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.lsClusterShowTerminated, "show-terminated", false, "when set, only show terminated clusters")
	cmd.Flags().StringVar(&r.args.lsClusterStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsClusterEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")
	cmd.Flags().BoolVarP(&r.args.lsClusterWatch, "watch", "w", false, "watch clusters")

	return cmd
}

func (r *runners) listClusters(_ *cobra.Command, args []string) error {
	const longForm = "2006-01-02T15:04:05Z"
	var startTime, endTime *time.Time
	if r.args.lsClusterStartTime != "" {
		st, err := time.Parse(longForm, r.args.lsClusterStartTime)
		if err != nil {
			return errors.Wrap(err, "parse start time")
		}
		startTime = &st
	}
	if r.args.lsClusterEndTime != "" {
		et, err := time.Parse(longForm, r.args.lsClusterEndTime)
		if err != nil {
			return errors.Wrap(err, "parse end time")
		}
		endTime = &et
	}

	clusters, err := r.kotsAPI.ListClusters(r.args.lsClusterShowTerminated, startTime, endTime)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "list clusters")
	}

	header := true
	if r.args.lsClusterWatch {

		// Checks to see if the outputFormat is table
		if r.outputFormat != "table" && r.outputFormat != "wide" {
			return errors.New("watch is only supported for table output")
		}

		clustersToPrint := make([]*types.Cluster, 0)

		// Prints the intial list of clusters
		if len(clusters) == 0 {
			print.NoClusters(r.outputFormat, r.w)
		} else {
			clustersToPrint = append(clustersToPrint, clusters...)
		}

		// Runs until ctrl C is recognized
		for range time.Tick(2 * time.Second) {
			newClusters, err := r.kotsAPI.ListClusters(r.args.lsClusterShowTerminated, startTime, endTime)
			if err != nil {
				if err == promptui.ErrInterrupt {
					return errors.New("interrupted")
				}

				return errors.Wrap(err, "watch clusters")
			}

			// Create a map from the IDs of the new clusters
			newClusterMap := make(map[string]*types.Cluster)
			for _, newCluster := range newClusters {
				newClusterMap[newCluster.ID] = newCluster
			}

			// Create a map from the IDs of the old clusters
			oldClusterMap := make(map[string]*types.Cluster)
			for _, cluster := range clusters {
				oldClusterMap[cluster.ID] = cluster
			}

			// Check for new clusters and print them
			for id, newCluster := range newClusterMap {
				if oldCluster, found := oldClusterMap[id]; !found {
					clustersToPrint = append(clustersToPrint, newCluster)
				} else {
					// Check if properties of existing clusters have changed
					// reset EstimatedCost (as it is calculated on the fly and not stored in the API response)
					oldCluster.EstimatedCost = 0
					if !reflect.DeepEqual(newCluster, oldCluster) {
						clustersToPrint = append(clustersToPrint, newCluster)
					}
				}
			}

			// Check for removed clusters and print them, changing their status to be "deleted"
			for id, cluster := range oldClusterMap {
				if _, found := newClusterMap[id]; !found {
					cluster.Status = types.ClusterStatusDeleted
					clustersToPrint = append(clustersToPrint, cluster)
				}
			}

			// Prints the clusters
			if len(clustersToPrint) > 0 {
				print.Clusters(r.outputFormat, r.w, clustersToPrint, header)
				header = false // only print the header once
			}

			clusters = newClusters
			clustersToPrint = make([]*types.Cluster, 0)
		}
	}

	if len(clusters) == 0 {
		return print.NoClusters(r.outputFormat, r.w)
	}

	return print.Clusters(r.outputFormat, r.w, clusters, true)
}
