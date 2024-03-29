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

	cmd.Flags().IntVar(&r.args.clusterExposePortPort, "port", 0, "Port to expose")
	cmd.MarkFlagRequired("port")

	cmd.Flags().StringArrayVar(&r.args.clusterExposePortProtocols, "protocol", []string{"http", "https"}, "Protocol to expose")
	cmd.MarkFlagRequired("protocol")

	return cmd
}

func (r *runners) clusterPortExpose(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	if len(r.args.clusterExposePortProtocols) == 0 {
		return errors.New("at least one protocol must be specified")
	}

	port, err := r.kotsAPI.ExposeClusterPort(clusterID, r.args.clusterExposePortPort, r.args.clusterExposePortProtocols)
	if err != nil {
		return err
	}

	if err := print.ClusterPort(r.outputFormat, r.w, port); err != nil {
		return err
	}

	return nil
}
