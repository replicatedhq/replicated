package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

// channelLsCmd represents the channelLs command
var channelLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all channels in your app",
	Long:  "List all channels in your app",
}

func init() {
	channelCmd.AddCommand(channelLsCmd)
}

func (r *runners) channelList(cmd *cobra.Command, args []string) error {
	channels, err := r.api.ListChannels(r.appID)
	if err != nil {
		return err
	}

	return print.Channels(r.w, channels)
}
