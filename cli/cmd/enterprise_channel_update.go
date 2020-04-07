package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannelUpdate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "update",
		SilenceUsage: true,
		Short:        "update an existing release channel",
		Long: `Update an existing shared release channel.

  Example:
  replicated enteprise channel update --id MyChannelID --name SomeBigBank --description 'The release channel for SomeBigBank'`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseChannelUpdateID, "id", "", "The id of the channel to be updated")
	cmd.Flags().StringVar(&r.args.enterpriseChannelUpdateName, "name", "", "The new name for this channel")
	cmd.Flags().StringVar(&r.args.enterpriseChannelUpdateDescription, "description", "", "The new description of this channel")

	cmd.RunE = r.enterpriseChannelUpdate
}

func (r *runners) enterpriseChannelUpdate(cmd *cobra.Command, args []string) error {
	channel, err := r.enterpriseClient.UpdateChannel(r.args.enterpriseChannelUpdateID, r.args.enterpriseChannelUpdateName, r.args.enterpriseChannelUpdateDescription)
	if err != nil {
		return err
	}

	print.EnterpriseChannel(r.w, channel)
	return nil
}
