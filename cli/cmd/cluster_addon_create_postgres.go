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

type clusterAddonCreatePostgresArgs struct {
	version      string
	diskGiB      int64
	instanceType string

	clusterID    string
	waitDuration time.Duration
	dryRun       bool
	outputFormat string
}

func (r *runners) InitClusterAddonCreatePostgres(parent *cobra.Command) *cobra.Command {
	args := clusterAddonCreatePostgresArgs{}

	cmd := &cobra.Command{
		Use:   "postgres CLUSTER_ID",
		Short: "Create a Postgres database for a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, cmdArgs []string) error {
			args.clusterID = cmdArgs[0]
			return r.clusterAddonCreatePostgresCreateRun(args)
		},
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	err := clusterAddonCreatePostgresFlags(cmd, &args)
	if err != nil {
		panic(err)
	}

	return cmd
}

func clusterAddonCreatePostgresFlags(cmd *cobra.Command, args *clusterAddonCreatePostgresArgs) error {
	cmd.Flags().StringVar(&args.version, "version", "", "The Postgres version to create")
	cmd.Flags().Int64Var(&args.diskGiB, "disk", 200, "Disk Size (GiB) for the Postgres database")
	cmd.Flags().StringVar(&args.instanceType, "instance-type", "db.t3.micro", "The type of instance to use for the Postgres database")

	cmd.Flags().DurationVar(&args.waitDuration, "wait", 0, "Wait duration for add-on to be ready before exiting (leave empty to not wait)")
	cmd.Flags().BoolVar(&args.dryRun, "dry-run", false, "Simulate creation to verify that your inputs are valid without actually creating an add-on")
	cmd.Flags().StringVar(&args.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")
	return nil
}

func (r *runners) clusterAddonCreatePostgresCreateRun(args clusterAddonCreatePostgresArgs) error {
	opts := kotsclient.CreateClusterAddonPostgresOpts{
		ClusterID:    args.clusterID,
		Version:      args.version,
		DiskGiB:      args.diskGiB,
		InstanceType: args.instanceType,
		DryRun:       args.dryRun,
	}

	addon, err := r.createAndWaitForClusterAddonCreatePostgres(opts, args.waitDuration)
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

func (r *runners) createAndWaitForClusterAddonCreatePostgres(opts kotsclient.CreateClusterAddonPostgresOpts, waitDuration time.Duration) (*types.ClusterAddon, error) {
	addon, err := r.kotsAPI.CreateClusterAddonPostgres(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster add-on postgres")
	}

	if opts.DryRun {
		return addon, nil
	}

	// if the wait flag was provided, we poll the api until the add-on is ready, or a timeout
	if waitDuration > 0 {
		return waitForAddon(r.kotsAPI, opts.ClusterID, addon.ID, waitDuration)
	}

	return addon, nil
}
