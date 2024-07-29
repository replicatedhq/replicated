package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ID [ID …]",
		Short: "Remove test clusters",
		Long: `Removes a cluster immediately.

You can specify the --all flag to terminate all clusters.`,
		RunE:              r.removeCluster,
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringArrayVar(&r.args.removeClusterNames, "name", []string{}, "Name of the cluster to remove (can be specified multiple times)")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)
	cmd.Flags().StringArrayVar(&r.args.removeClusterTags, "tag", []string{}, "Tag of the cluster to remove (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.removeClusterAll, "all", false, "remove all clusters")

	cmd.Flags().BoolVar(&r.args.removeClusterDryRun, "dry-run", false, "Dry run")

	return cmd
}

func (r *runners) removeCluster(_ *cobra.Command, args []string) error {
	if len(args) == 0 && !r.args.removeClusterAll && len(r.args.removeClusterNames) == 0 && len(r.args.removeClusterTags) == 0 {
		return errors.New("One of ID, --all, --name or --tag flag required")
	} else if len(args) > 0 && (r.args.removeClusterAll || len(r.args.removeClusterNames) > 0 || len(r.args.removeClusterTags) > 0) {
		return errors.New("cannot specify ID and --all, --name or --tag flag")
	} else if len(args) == 0 && r.args.removeClusterAll && (len(r.args.removeClusterNames) > 0 || len(r.args.removeClusterTags) > 0) {
		return errors.New("cannot specify --all and --name or --tag flag")
	} else if len(args) == 0 && !r.args.removeClusterAll && len(r.args.removeClusterNames) > 0 && len(r.args.removeClusterTags) > 0 {
		return errors.New("cannot specify --name and --tag flag")
	}

	if len(r.args.removeClusterNames) > 0 {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			for _, name := range r.args.removeClusterNames {
				if cluster.Name == name {
					err := remove(r, cluster.ID)
					if err != nil {
						return errors.Wrap(err, "remove cluster")
					}
				}
			}
		}
	}

	if len(r.args.removeClusterTags) > 0 {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		tags, err := parseTags(r.args.removeClusterTags)
		if err != nil {
			return errors.Wrap(err, "parse tags")
		}

		for _, cluster := range clusters {
			if cluster.Tags != nil && len(cluster.Tags) > 0 {
				for _, tag := range tags {
					for _, clusterTag := range cluster.Tags {
						if clusterTag.Key == tag.Key && clusterTag.Value == tag.Value {
							err := remove(r, cluster.ID)
							if err != nil {
								return errors.Wrap(err, "remove cluster")
							}
						}
					}
				}
			}
		}
	}

	if r.args.removeClusterAll {
		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}
		for _, cluster := range clusters {
			err := remove(r, cluster.ID)
			if err != nil {
				return errors.Wrap(err, "remove cluster")
			}
		}
	}

	for _, arg := range args {
		err := remove(r, arg)
		if err != nil {
			return errors.Wrap(err, "remove cluster")
		}
	}

	return nil
}

func remove(r *runners, clusterID string) error {
	if r.args.removeClusterDryRun {
		fmt.Printf("would remove cluster %s\n", clusterID)
		return nil
	}
	err := r.kotsAPI.RemoveCluster(clusterID)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return errors.Wrap(err, "remove cluster")
	} else {
		fmt.Printf("removed cluster %s\n", clusterID)
	}
	return nil
}
