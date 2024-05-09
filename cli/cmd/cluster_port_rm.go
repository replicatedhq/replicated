package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "rm CLUSTER_ID --id PORT_ID",
		RunE: r.clusterPortRemove,
		Args: cobra.ExactArgs(1),
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.clusterPortRemoveAddonID, "id", "", "ID of the port to remove (required)")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table|wide (default: table)")

	// Deprecated flags
	cmd.Flags().IntVar(&r.args.clusterPortRemovePort, "port", 0, "Port to remove")
	err := cmd.Flags().MarkHidden("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringArrayVar(&r.args.clusterPortRemoveProtocols, "protocol", []string{"http", "https"}, "Protocol to remove")
	err = cmd.Flags().MarkHidden("protocol")
	if err != nil {
		panic(err)
	}

	return cmd
}

func (r *runners) clusterPortRemove(_ *cobra.Command, args []string) error {
	clusterID := args[0]

	if r.args.clusterPortRemoveAddonID == "" && r.args.clusterPortRemovePort == 0 {
		return errors.New("either --id or --port must be specified")
	} else if r.args.clusterPortRemoveAddonID != "" && r.args.clusterPortRemovePort > 0 {
		return errors.New("only one o  --id or --port can be specified")
	}

	if r.args.clusterPortRemoveAddonID != "" {
		err := r.kotsAPI.DeleteClusterAddon(clusterID, r.args.clusterPortRemoveAddonID)
		if err != nil {
			return err
		}

		ports, err := r.kotsAPI.ListClusterPorts(clusterID)
		if err != nil {
			return err
		}

		return print.ClusterPorts(r.outputFormat, r.w, ports, true)
	}

	if len(r.args.clusterPortRemoveProtocols) == 0 {
		return errors.New("at least one protocol must be specified")
	}

	ports, err := r.kotsAPI.RemoveClusterPort(clusterID, r.args.clusterPortRemovePort, r.args.clusterPortRemoveProtocols)
	if err != nil {
		return err
	}

	return print.ClusterPorts(r.outputFormat, r.w, ports, true)
}
