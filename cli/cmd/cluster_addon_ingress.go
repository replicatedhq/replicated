package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitClusterAddOnIngress(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ingress",
	}
	parent.AddCommand(cmd)

	return cmd
}
