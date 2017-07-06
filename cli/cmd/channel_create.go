package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new channel in your app",
	Long: `Create a new channel in your app and print the full set of channels in the app on success.

Example:
replicated channel create --name Beta --description 'New features subject to change'`,
}

var channelCreateName string
var channelCreateDescription string

func init() {
	channelCmd.AddCommand(channelCreateCmd)

	channelCreateCmd.Flags().StringVar(&channelCreateName, "name", "", "The name of this channel")
	channelCreateCmd.Flags().StringVar(&channelCreateDescription, "description", "", "A longer description of this channel")
}

func (r *runners) channelCreate(cmd *cobra.Command, args []string) error {
	allChannels, err := r.api.CreateChannel(r.appID, channelCreateName, channelCreateDescription)
	if err != nil {
		return err
	}

	return print.Channels(r.w, allChannels)
}
