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
		Use:  "create <cluster_id> --target svc/my-service --port <port>",
		RunE: r.ingressClusterCreate,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.clusterCreateIngressTarget, "target", "", "The target for the ingress")
	cmd.MarkFlagRequired("target")

	cmd.Flags().IntVar(&r.args.clusterCreateIngressPort, "port", 80, "The port for the ingress")
	cmd.MarkFlagRequired("port")

	cmd.Flags().StringVar(&r.args.clusterCreateIngressNamespace, "namespace", "default", "The namespace for the ingress")

	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) ingressClusterCreate(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	namespace := r.args.clusterCreateIngressNamespace
	if namespace == "" {
		namespace = "default" // avoiding the entire k8s dep list
	}

	opts := kotsclient.CreateClusterIngressOpts{
		ClusterID: clusterID,
		Target:    r.args.clusterCreateIngressTarget,
		Port:      r.args.clusterCreateIngressPort,
		Namespace: namespace,
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
