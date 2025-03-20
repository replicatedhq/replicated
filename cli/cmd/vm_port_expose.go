package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitVMPortExpose(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expose VM_ID --port PORT",
		Short: "Expose a port on a vm to the public internet.",
		Long: `The 'vm port expose' command is used to expose a specified port on a vm to the public internet. When exposing a port, the command automatically creates a DNS entry and, if using the "https" protocol, provisions a TLS certificate for secure communication.

You can also create a wildcard DNS entry and TLS certificate by specifying the "--wildcard" flag. Please note that creating a wildcard certificate may take additional time.

This command supports different protocols including "http", "https", "ws", and "wss" for web traffic and web socket communication.`,
		Example: `# Expose port 8080 with HTTPS protocol and wildcard DNS
replicated vm port expose VM_ID --port 8080 --protocol https --wildcard

# Expose port 3000 with HTTP protocol
replicated vm port expose VM_ID --port 3000 --protocol http

# Expose port 8080 with multiple protocols
replicated vm port expose VM_ID --port 8080 --protocol http,https

# Expose port 8080 and display the result in JSON format
replicated vm port expose VM_ID --port 8080 --protocol https --output json`,
		RunE:              r.vmPortExpose,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: r.completeVMIDs,
	}
	parent.AddCommand(cmd)

	cmd.Flags().IntVar(&r.args.vmExposePortPort, "port", 0, "Port to expose (required)")
	err := cmd.MarkFlagRequired("port")
	if err != nil {
		panic(err)
	}
	cmd.Flags().StringSliceVar(&r.args.vmExposePortProtocols, "protocol", []string{"http", "https"}, `Protocol to expose (valid values are "http", "https", "ws" and "wss")`)
	cmd.Flags().BoolVar(&r.args.vmExposePortIsWildcard, "wildcard", false, "Create a wildcard DNS entry and TLS certificate for this port")
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table|wide")

	return cmd
}

func (r *runners) vmPortExpose(_ *cobra.Command, args []string) error {
	vmID := args[0]

	if len(r.args.vmExposePortProtocols) == 0 {
		return errors.New("at least one protocol must be specified")
	}

	port, err := r.kotsAPI.ExposeVMPort(
		vmID,
		r.args.vmExposePortPort, r.args.vmExposePortProtocols, r.args.vmExposePortIsWildcard,
	)
	if err != nil {
		return err
	}

	return print.VMPort(r.outputFormat, r.w, port)
}
