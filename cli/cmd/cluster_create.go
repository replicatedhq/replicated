package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create test clusters",
		Long:         `create test clusters`,
		RunE:         r.createCluster,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createClusterName, "name", "", "cluster name")
	cmd.Flags().StringVar(&r.args.createClusterDistribution, "distribution", "", "cluster distribution")
	cmd.Flags().StringVar(&r.args.createClusterVersion, "version", "", "cluster version")
	cmd.Flags().StringVar(&r.args.createClusterTTL, "ttl", "", "cluster ttl (duration)")

	return cmd
}

func (r *runners) createCluster(_ *cobra.Command, args []string) error {
	kotsRestClient := kotsclient.VendorV3Client{HTTPClient: *r.platformAPI}

	opts := kotsclient.CreateClusterOpts{
		Name:         r.args.createClusterName,
		Distribution: r.args.createClusterDistribution,
		Version:      r.args.createClusterVersion,
		TTL:          r.args.createClusterTTL,
	}
	_, err := kotsRestClient.CreateCluster(opts)
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}
