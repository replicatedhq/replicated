package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "rm",
		RunE: r.clusterPortRemove,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().IntVar(&r.args.clusterPortRemovePort, "port", 0, "Port to remove")
	cmd.Flags().StringArrayVar(&r.args.clusterPortRemoveProtocols, "protocol", []string{"http"}, "Protocol to remove")

	return cmd
}

func (r *runners) clusterPortRemove(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	if len(r.args.clusterPortRemoveProtocols) == 0 {
		return errors.New("at least one protocol must be specified")
	}

	ports, err := r.kotsAPI.RemoveClusterPort(clusterID, r.args.clusterPortRemovePort, r.args.clusterPortRemoveProtocols)
	if err != nil {
		return err
	}

	print.ClusterPorts(r.outputFormat, r.w, ports, true)
	return nil
}
