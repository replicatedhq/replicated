package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

const (
	clusterPortExposeShort = "Expose a port on a cluster to the public internet"
	clusterPortExposeLong  = `Expose a port on a cluster to the public internet.

This command will create a DNS entry and TLS certificate (if "https") for the specified port on the cluster.

A wildcard DNS entry and TLS certificate can be created by specifying the "--wildcard" flag. This will take extra time to provision.

NOTE: This feature currently only supports VM cluster distributions.`
	clusterPortExposeExample = `  $ replicated cluster port expose 05929b24 --port 8080 --protocol https --wildcard
  ID              CLUSTER PORT    PROTOCOL        EXPOSED PORT                                           WILDCARD        STATUS
  d079b2fc        8080            https           https://happy-germain.ingress.replicatedcluster.com    true            pending`
)

func (r *runners) InitClusterPortExpose(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "expose CLUSTER_ID --port PORT",
		Short:             clusterPortExposeShort,
		Long:              clusterPortExposeLong,
		Example:           clusterPortExposeExample,
		RunE:              r.clusterPortExpose,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeClusterIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().IntVar(&r.args.clusterExposePortPort, "port", 0, "Port to expose (required)")
	err := cmd.MarkFlagRequired("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringSliceVar(&r.args.clusterExposePortProtocols, "protocol", []string{"http", "https"}, `Protocol to expose (valid values are "http", "https", "ws" and "wss")`)
	cmd.Flags().BoolVar(&r.args.clusterExposePortIsWildcard, "wildcard", false, "Create a wildcard DNS entry and TLS certificate for this port")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	return cmd
}

func (r *runners) clusterPortExpose(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	if len(r.args.clusterExposePortProtocols) == 0 {
		return errors.New("at least one protocol must be specified")
	}

	port, err := r.kotsAPI.ExposeClusterPort(
		clusterID,
		r.args.clusterExposePortPort, r.args.clusterExposePortProtocols, r.args.clusterExposePortIsWildcard,
	)
	if err != nil {
		return err
	}

	return print.ClusterPort(r.outputFormat, r.w, port)
}
