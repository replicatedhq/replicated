package cmd

import (
	"fmt"
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
		Use:          "ls",
		Short:        "list test clusters",
		Long:         `list test clusters`,
		RunE:         r.listClusters,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().BoolVar(&r.args.lsClusterShowTerminated, "show-terminated", false, "when set, only show terminated clusters")
	cmd.Flags().StringVar(&r.args.lsClusterStartTime, "start-time", "", "start time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.args.lsClusterEndTime, "end-time", "", "end time for the query (Format: 2006-01-02T15:04:05Z)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")
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

	if r.args.lsClusterWatch {

		// Checks to see if the outputFormat is table
		if r.outputFormat != "table" {
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
					if !reflect.DeepEqual(newCluster, oldCluster) {
						clustersToPrint = append(clustersToPrint, newCluster)
					}
				}
			}

			// Check for removed clusters and print them, changing their status to be "deleted"
			for id, cluster := range oldClusterMap {
				if _, found := newClusterMap[id]; !found {
					cluster.Status = "deleted"
					clustersToPrint = append(clustersToPrint, cluster)
				}
			}

			// Prints the clusters
			if len(clustersToPrint) > 0 {
				fmt.Print("\033[H\033[2J") // Clears the console
				print.Clusters(r.outputFormat, r.w, clustersToPrint)
			}

			clusters = newClusters
		}
	}

	if len(clusters) == 0 {
		return print.NoClusters(r.outputFormat, r.w)
	}

	return print.Clusters(r.outputFormat, r.w, clusters)
}
