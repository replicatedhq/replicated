package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannelCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:          "create",
		SilenceUsage: true,
		Short:        "Create a new shared release channel",
		Long: `Create a new shared release channel for vendors.

  Example:
  replicated enteprise channel create --name SomeBigBank --description 'The release channel for SomeBigBank'`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.enterpriseChannelCreateName, "name", "", "The name of this channel")
	cmd.Flags().StringVar(&r.args.enterpriseChannelCreateDescription, "description", "", "A longer description of this channel")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.enterpriseChannelCreate
}

func (r *runners) enterpriseChannelCreate(cmd *cobra.Command, args []string) error {
	channel, err := r.enterpriseClient.CreateChannel(r.args.enterpriseChannelCreateName, r.args.enterpriseChannelCreateDescription)
	if err != nil {
		return err
	}

	print.EnterpriseChannel(r.outputFormat, r.w, channel)
	return nil
}
