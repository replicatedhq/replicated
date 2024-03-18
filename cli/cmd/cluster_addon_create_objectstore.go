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

type clusterAddonCreateObjectStoreArgs struct {
	objectStoreBucket string
	clusterAddonCreateArgs
	clusterAddonArgs
}

func (r *runners) InitClusterAddonCreateObjectStore(parent *cobra.Command) *cobra.Command {
	args := clusterAddonCreateObjectStoreArgs{}

	cmd := &cobra.Command{
		Use:   "object-store CLUSTER_ID --bucket BUCKET_NAME",
		Short: "Create an object store for a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.clusterAddonCreateObjectStoreCreateRun(args)
		},
	}
	parent.AddCommand(cmd)

	_ = clusterAddonCreateObjectStoreFlags(cmd, &args)

	return cmd
}

func clusterAddonCreateObjectStoreFlags(cmd *cobra.Command, args *clusterAddonCreateObjectStoreArgs) error {
	cmd.Flags().StringVar(&args.objectStoreBucket, "bucket", "", "The object store bucket name to create (required)")
	err := cmd.MarkFlagRequired("bucket")
	if err != nil {
		return err
	}
	err = clusterAddonFlags(cmd, &args.clusterAddonArgs)
	if err != nil {
		return err
	}
	return clusterAddonCreateFlags(cmd, &args.clusterAddonCreateArgs)
}

func (r *runners) clusterAddonCreateObjectStoreCreateRun(args clusterAddonCreateObjectStoreArgs) error {
	opts := kotsclient.CreateClusterAddonObjectStoreOpts{
		ClusterID: args.clusterID,
		Bucket:    args.objectStoreBucket,
		DryRun:    args.dryRun,
	}

	addon, err := r.createAndWaitForClusterAddonCreateObjectStore(opts, args.waitDuration)
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
		_, err := fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.Addon(args.outputFormat, r.w, addon)
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

	// if the wait flag was provided, we poll the api until the cluster is ready, or a timeout
	if waitDuration > 0 {
		return waitForAddon(r.kotsAPI, opts.ClusterID, addon.ID, waitDuration)
	}

	return addon, nil
}
