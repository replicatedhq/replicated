package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannelAssign(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "assign",
		SilenceUsage: true,
		Short:        "assign a channel to a vendor",
		Long: `Assign a channel to a vendor team.

  Example:
  replicated enteprise channel assign --channel-id ChannelID --vendor-id VendorID`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseChannelAssignChannelID, "channel-id", "", "The id of the channel to be assigned")
	cmd.Flags().StringVar(&r.args.enterpriseChannelAssignVendorID, "vendor-id", "", "The id of the vendor to assign the channel to")

	cmd.RunE = r.enterpriseChannelAssign
}

func (r *runners) enterpriseChannelAssign(cmd *cobra.Command, args []string) error {
	err := r.enterpriseClient.AssignChannel(r.args.enterpriseChannelAssignChannelID, r.args.enterpriseChannelAssignVendorID)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Channel successfully assigned\n")
	r.w.Flush()

	return nil
}
