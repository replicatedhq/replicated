package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/moby/moby/pkg/namesgenerator"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitNetworkCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create networks for VMs and Clusters",
		Long:         ``,
		SilenceUsage: true,
		RunE:         r.createNetwork,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createNetworkName, "name", "", "Network name (defaults to random name)")
	cmd.Flags().StringVar(&r.args.createNetworkVersion, "version", "v1", "Network version to use: v1 (default) or v2 (alpha)")
	cmd.Flags().MarkHidden("version")
	cmd.Flags().StringVar(&r.args.createNetworkTTL, "ttl", "", "Network TTL (duration, max 48h)")
	cmd.Flags().DurationVar(&r.args.createNetworkWaitDuration, "wait", time.Second*0, "Wait duration for Network to be ready (leave empty to not wait)")

	cmd.Flags().BoolVar(&r.args.createNetworkDryRun, "dry-run", false, "Dry run")

	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) createNetwork(_ *cobra.Command, args []string) error {
	if r.args.createNetworkName == "" {
		r.args.createNetworkName = namesgenerator.GetRandomName(0)
	}

	opts := kotsclient.CreateNetworkOpts{
		Name:    r.args.createNetworkName,
		Version: r.args.createNetworkVersion,
		TTL:     r.args.createNetworkTTL,
		DryRun:  r.args.createNetworkDryRun,
	}

	network, err := r.createAndWaitForNetwork(opts)
	if err != nil {
		if errors.Cause(err) == ErrVMWaitDurationExceeded {
			defer os.Exit(124)
		} else {
			return err
		}
	}

	if opts.DryRun {
		_, err = fmt.Fprintln(r.w, "Dry run succeeded.")
		return err
	}

	return print.Network(r.outputFormat, r.w, network)
}

func (r *runners) createAndWaitForNetwork(opts kotsclient.CreateNetworkOpts) (*types.Network, error) {
	network, ve, err := r.kotsAPI.CreateNetwork(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create network")
	}

	if ve != nil && ve.Message != "" {
		return nil, errors.New(ve.Message)
	}

	if opts.DryRun {
		return network, nil
	}

	// if the wait flag was provided, we poll the api until the network is ready, or a timeout
	if r.args.createNetworkWaitDuration > 0 {
		return waitForNetwork(r.kotsAPI, network, r.args.createNetworkWaitDuration)
	}

	return network, nil
}

func waitForNetwork(kotsRestClient *kotsclient.VendorV3Client, network *types.Network, duration time.Duration) (*types.Network, error) {
	start := time.Now()
	for {
		network, err := kotsRestClient.GetNetwork(network.ID)
		if err != nil {
			return nil, errors.Wrap(err, "get network")
		}

		if network.Status == types.NetworkStatusRunning {
			return network, nil
		}
		if network.Status == types.NetworkStatusError {
			return nil, errors.New("network failed to provision")
		}
		if time.Now().After(start.Add(duration)) {
			// In case of timeout, return the network and a WaitDurationExceeded error
			return network, ErrWaitDurationExceeded
		}

		time.Sleep(time.Second * 5)
	}
}
