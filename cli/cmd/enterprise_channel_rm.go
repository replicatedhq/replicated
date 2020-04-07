package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannelRM(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "rm",
		SilenceUsage: true,
		Short:        "Remove a channel",
		Long: `Remove a channel.

  Example:
  replicated enteprise channel rm --id MyChannelID`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseChannelRmId, "id", "", "The id of the channel to remove")

	cmd.RunE = r.enterpriseChannelRemove
}

func (r *runners) enterpriseChannelRemove(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.RemoveChannel(r.args.enterpriseChannelRmId)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Channel %s successfully removed\n", r.args.enterpriseChannelRmId)
	r.w.Flush()

	return nil
}
