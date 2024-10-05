package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterPort(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "port",
		Short: "Manage cluster ports.",
		Long: `The 'cluster port' command is a parent command for managing ports in a cluster. It allows users to list, remove, or expose specific ports used by the cluster. Use the subcommands (such as 'ls', 'rm', and 'expose') to manage port configurations effectively.

This command provides flexibility for handling ports in various test clusters, ensuring efficient management of cluster networking settings.`,
		Example: `  # List all exposed ports in a cluster
  replicated cluster port ls [CLUSTER_ID]

  # Remove an exposed port from a cluster
  replicated cluster port rm [CLUSTER_ID] [PORT]

  # Expose a new port in a cluster
  replicated cluster port expose [CLUSTER_ID] [PORT]`,
		SilenceUsage: true,
		Hidden:       false,
	}
	parent.AddCommand(cmd)

	return cmd
}
