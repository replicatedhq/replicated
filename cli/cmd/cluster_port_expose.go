package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPortExpose(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expose CLUSTER_ID_OR_NAME --port PORT",
		Short: "Expose a port on a cluster to the public internet.",
		Long: `The 'cluster port expose' command is used to expose a specified port on a cluster to the public internet. When exposing a port, the command automatically creates a DNS entry and, if using the "https" protocol, provisions a TLS certificate for secure communication.

You can also create a wildcard DNS entry and TLS certificate by specifying the "--wildcard" flag. Please note that creating a wildcard certificate may take additional time.

This command supports different protocols including "http", "https", "ws", and "wss" for web traffic and web socket communication.

NOTE: Currently, this feature only supports VM-based cluster distributions.`,
		Example: `# Expose port 8080 with HTTPS protocol and wildcard DNS
replicated cluster port expose CLUSTER_ID_OR_NAME --port 8080 --protocol https --wildcard

# Expose port 30000 with HTTP protocol
replicated cluster port expose CLUSTER_ID_OR_NAME --port 30000 --protocol http

# Expose port 8080 with multiple protocols
replicated cluster port expose CLUSTER_ID_OR_NAME --port 8080 --protocol http,https

# Expose port 8080 and display the result in JSON format
replicated cluster port expose CLUSTER_ID_OR_NAME --port 8080 --protocol https --output json`,
		RunE:              r.clusterPortExpose,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeClusterIDsAndNames,
	}
	parent.AddCommand(cmd)

	cmd.Flags().IntVar(&r.args.clusterExposePortPort, "port", 0, "Port to expose (required)")
	err := cmd.MarkFlagRequired("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringSliceVar(&r.args.clusterExposePortProtocols, "protocol", []string{"http", "https"}, `Protocol to expose (valid values are "http", "https", "ws" and "wss")`)
	cmd.Flags().BoolVar(&r.args.clusterExposePortIsWildcard, "wildcard", false, "Create a wildcard DNS entry and TLS certificate for this port")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) clusterPortExpose(_ *cobra.Command, args []string) error {
	clusterID, err := r.getClusterIDFromArg(args[0])
	if err != nil {
		return err
	}

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
