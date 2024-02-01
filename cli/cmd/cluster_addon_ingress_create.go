package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/replicatedhq/replicated/pkg/platformclient"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddOnIngressCreate(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create",
		RunE: r.ingressClusterCreate,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) ingressClusterCreate(_ *cobra.Command, args []string) error {
	// for consistency in the cli, these are positional args
	if len(args) != 2 {
		return errors.New("cluster ingress create requires exactly two arguments: cluster id and ingress target")
	}

	clusterID := args[0]
	ingressTarget := args[1]

	opts := kotsclient.CreateClusterIngressOpts{
		ClusterID: clusterID,
		Target:    ingressTarget,
	}

	ing, err := r.createAndWaitForIngress(opts)
	if err != nil {
		if errors.Cause(err) == ErrWaitDurationExceeded {
			defer func() {
				os.Exit(124)
			}()
		} else {
			return err
		}
	}

	return print.AddOn(r.outputFormat, r.w, ing)
}

func (r *runners) createAndWaitForIngress(opts kotsclient.CreateClusterIngressOpts) (*types.ClusterAddOn, error) {
	ing, err := r.kotsAPI.CreateClusterIngress(opts)
	if errors.Cause(err) == platformclient.ErrForbidden {
		return nil, ErrCompatibilityMatrixTermsNotAccepted
	} else if err != nil {
		return nil, errors.Wrap(err, "create cluster ingress")
	}

	return ing, nil
}
