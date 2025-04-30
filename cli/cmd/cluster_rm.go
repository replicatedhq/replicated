package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterRemove(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm ID_OR_NAME [ID_OR_NAME …]",
		Aliases: []string{"delete"},
		Short:   "Remove test clusters.",
		Long: `The 'rm' command removes test clusters immediately.

You can remove clusters by specifying a cluster ID or name, or by using other criteria such as cluster tags. Alternatively, you can remove all clusters in your account at once.

When specifying a name that matches multiple clusters, all clusters with that name will be removed.

This command can also be used in a dry-run mode to simulate the removal without actually deleting anything.

You cannot mix the use of cluster IDs or names with other options like removing by tag or removing all clusters at once.`,
		Example: `# Remove a specific cluster by ID or name
replicated cluster rm CLUSTER_ID_OR_NAME

# Remove multiple clusters by ID or name
replicated cluster rm CLUSTER_ID_1 CLUSTER_NAME_2

# Remove all clusters
replicated cluster rm --all`,
		RunE:              r.removeClusters,
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringArrayVar(&r.args.removeClusterNames, "name", []string{}, "Name of the cluster to remove (can be specified multiple times) (deprecated: use ID_OR_NAME arguments instead)")
	cmd.RegisterFlagCompletionFunc("name", r.completeClusterNames)
	cmd.Flag("name").Deprecated = "use ID_OR_NAME arguments instead"
	cmd.Flags().StringArrayVar(&r.args.removeClusterTags, "tag", []string{}, "Tag of the cluster to remove (key=value format, can be specified multiple times)")

	cmd.Flags().BoolVar(&r.args.removeClusterAll, "all", false, "remove all clusters")

	cmd.Flags().BoolVar(&r.args.removeClusterDryRun, "dry-run", false, "Dry run")

	return cmd
}

func (r *runners) removeClusters(_ *cobra.Command, args []string) error {
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
					err := removeCuster(r, cluster.ID)
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
			if len(cluster.Tags) > 0 {
				for _, tag := range tags {
					for _, clusterTag := range cluster.Tags {
						if clusterTag.Key == tag.Key && clusterTag.Value == tag.Value {
							err := removeCuster(r, cluster.ID)
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
			err := removeCuster(r, cluster.ID)
			if err != nil {
				return errors.Wrap(err, "remove cluster")
			}
		}
	}

	for _, arg := range args {
		_, err := r.kotsAPI.GetCluster(arg)
		if err == nil {
			err := removeCuster(r, arg)
			if err != nil {
				return errors.Wrap(err, "remove cluster")
			}
			continue
		}

		clusters, err := r.kotsAPI.ListClusters(false, nil, nil)
		if err != nil {
			return errors.Wrap(err, "list clusters")
		}

		found := false
		for _, cluster := range clusters {
			if cluster.Name == arg {
				found = true
				err := removeCuster(r, cluster.ID)
				if err != nil {
					return errors.Wrap(err, "remove cluster")
				}
			}
		}

		if !found {
			return errors.Errorf("Cluster with name or ID '%s' not found", arg)
		}
	}

	return nil
}

func removeCuster(r *runners, clusterID string) error {
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
