package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelCreate(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new channel in your app",
		Long: `Create a new channel in your app and print the channel on success.

  Example:
  replicated channel create --name Beta --description 'New features subject to change'`,
	}
	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.channelCreateName, "name", "", "The name of this channel")
	cmd.Flags().StringVar(&r.args.channelCreateDescription, "description", "", "A longer description of this channel")
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.channelCreate
}

func (r *runners) channelCreate(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	allChannels, err := r.api.CreateChannel(r.appID, r.appType, r.args.channelCreateName, r.args.channelCreateDescription)
	if err != nil {
		return err
	}

	return print.Channels(r.outputFormat, r.w, allChannels)
}
