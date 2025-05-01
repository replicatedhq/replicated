package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortRm(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm CLUSTER_ID_OR_NAME --id PORT_ID",
		Aliases: []string{"delete"},
		Short:   "Remove cluster port by ID.",
		Long: `The 'cluster port rm' command removes a specific port from a cluster. You must provide the ID or name of the cluster and either the ID of the port or the port number and protocol(s) to remove.

This command is useful for managing the network settings of your test clusters by allowing you to clean up unused or incorrect ports. After removing a port, the updated list of ports will be displayed.

Note that you can only use either the port ID or port number when removing a port, not both at the same time.`,
		Example: `# Remove a port using its ID
replicated cluster port rm CLUSTER_ID_OR_NAME --id PORT_ID

# Remove a port using its number (deprecated)
replicated cluster port rm CLUSTER_ID_OR_NAME --port 8080 --protocol http,https

# Remove a port and display the result in JSON format
replicated cluster port rm CLUSTER_ID_OR_NAME --id PORT_ID --output json`,
		RunE:              r.clusterPortRemove,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.clusterPortRemoveAddonID, "id", "", "ID of the port to remove (required)")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	// Deprecated flags
	cmd.Flags().IntVar(&r.args.clusterPortRemovePort, "port", 0, "Port to remove")
	err := cmd.Flags().MarkHidden("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringSliceVar(&r.args.clusterPortRemoveProtocols, "protocol", []string{"http", "https"}, "Protocol to remove")
	err = cmd.Flags().MarkHidden("protocol")
	if err != nil {
		panic(err)
	}

	return cmd
}

func (r *runners) clusterPortRemove(_ *cobra.Command, args []string) error {
	clusterID, err := r.getClusterIDFromArg(args[0])
	if err != nil {
		return err
	}

	if r.args.clusterPortRemoveAddonID == "" && r.args.clusterPortRemovePort == 0 {
		return errors.New("either --id or --port must be specified")
	} else if r.args.clusterPortRemoveAddonID != "" && r.args.clusterPortRemovePort > 0 {
		return errors.New("only one of --id or --port can be specified")
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
