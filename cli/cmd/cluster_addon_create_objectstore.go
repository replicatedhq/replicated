package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddonCreateObjectStore(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "object-store CLUSTER_ID --bucket-prefix BUCKET_PREFIX",
		Short: "Create an object store bucket for a cluster.",
		Long:  `Creates an object store bucket for a cluster, requiring a bucket name prefix. The bucket name will be auto-generated using the format "[BUCKET_PREFIX]-[ADDON_ID]-cmx". This feature provisions an object storage bucket that can be used for storage in your cluster environment.`,
		Example: `# Create an object store bucket with a specified prefix
replicated cluster addon create object-store 05929b24 --bucket-prefix mybucket

# Create an object store bucket and wait for it to be ready (up to 5 minutes)
replicated cluster addon create object-store 05929b24 --bucket-prefix mybucket --wait 5m

# Perform a dry run to validate inputs without creating the bucket
replicated cluster addon create object-store 05929b24 --bucket-prefix mybucket --dry-run

# Create an object store bucket and output the result in JSON format
replicated cluster addon create object-store 05929b24 --bucket-prefix mybucket --output json

# Create an object store bucket with a custom prefix and wait for 10 minutes
replicated cluster addon create object-store 05929b24 --bucket-prefix custom-prefix --wait 10m`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			r.args.clusterAddonCreateObjectStoreClusterID = cmdArgs[0]
			return r.clusterAddonCreateObjectStoreCreateRun()
		},
	}
	parent.AddCommand(cmd)

	err := r.clusterAddonCreateObjectStoreFlags(cmd)
	if err != nil {
		panic(err)
	}

	return cmd
}

func (r *runners) clusterAddonCreateObjectStoreFlags(cmd *cobra.Command) error {
	cmd.Flags().StringVar(&r.args.clusterAddonCreateObjectStoreBucket, "bucket-prefix", "", "A prefix for the bucket name to be created (required)")
	err := cmd.MarkFlagRequired("bucket-prefix")
	if err != nil {
		return err
	}
	cmd.Flags().DurationVar(&r.args.clusterAddonCreateObjectStoreDuration, "wait", 0, "Wait duration for add-on to be ready before exiting (leave empty to not wait)")
	cmd.Flags().BoolVar(&r.args.clusterAddonCreateObjectStoreDryRun, "dry-run", false, "Simulate creation to verify that your inputs are valid without actually creating an add-on")
	cmd.Flags().StringVar(&r.args.clusterAddonCreateObjectStoreOutput, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	return nil
}

func (r *runners) clusterAddonCreateObjectStoreCreateRun() error {
	opts := kotsclient.CreateClusterAddonObjectStoreOpts{
		ClusterID: r.args.clusterAddonCreateObjectStoreClusterID,
		Bucket:    r.args.clusterAddonCreateObjectStoreBucket,
		DryRun:    r.args.clusterAddonCreateObjectStoreDryRun,
	}

	addon, err := r.createAndWaitForClusterAddonCreateObjectStore(opts, r.args.clusterAddonCreateObjectStoreDuration)
	if err != nil {
		if errors.Cause(err) == ErrWaitDurationExceeded {
			defer func() {
				os.Exit(124)
			}()
		} else {
			return err
		}
	}

	if opts.DryRun {
		_, err := fmt.Fprintln(r.w, "Dry run succeeded for addon object-store creation.")
		return err
	}

	return print.Addon(r.args.clusterAddonCreateObjectStoreOutput, r.w, addon)
}

func (r *runners) createAndWaitForClusterAddonCreateObjectStore(opts kotsclient.CreateClusterAddonObjectStoreOpts, waitDuration time.Duration) (*types.ClusterAddon, error) {
	addon, err := r.kotsAPI.CreateClusterAddonObjectStore(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster addon object store")
	}

	if opts.DryRun {
		return addon, nil
	}

	// if the wait flag was provided, we poll the api until the addon is ready, or a timeout
	if waitDuration > 0 {
		return waitForAddon(r.kotsAPI, opts.ClusterID, addon.ID, waitDuration)
	}

	return addon, nil
}
