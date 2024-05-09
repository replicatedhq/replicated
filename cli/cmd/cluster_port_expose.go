package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortExpose(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "expose CLUSTER_ID --port PORT --protocol PROTOCOL",
		RunE: r.clusterPortExpose,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().IntVar(&r.args.clusterExposePortPort, "port", 0, "Port to expose (required)")
	err := cmd.MarkFlagRequired("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringArrayVar(&r.args.clusterExposePortProtocols, "protocol", []string{"http", "https"}, "Protocol to expose")
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
