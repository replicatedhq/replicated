package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannel(parent *cobra.Command) *cobra.Command {
	enterpriseChannelCommand := &cobra.Command{
		Use:   "channel",
		Short: "Manage enterprise channels",
		Long:  `The channel command allows approved enterprise to create custom release channels`,
	}
	parent.AddCommand(enterpriseChannelCommand)

	return enterpriseChannelCommand
}
