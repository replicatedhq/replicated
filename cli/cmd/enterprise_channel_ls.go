package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseChannelLS(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ls",
		Short:        "lists enterprise channels",
		Long:         `lists all channels that have been created`,
		RunE:         r.enterpriseChannelList,
		SilenceUsage: true,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	return cmd
}

func (r *runners) enterpriseChannelList(cmd *cobra.Command, args []string) error {
	channels, err := r.enterpriseClient.ListChannels()
	if err != nil {
		return err
	}

	print.EnterpriseChannels(r.outputFormat, r.w, channels)
	return nil
}
