package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
)

// channelInspectCmd represents the channelInspect command
var channelInspectCmd = &cobra.Command{
	Use:   "inspect CHANNEL_ID",
	Short: "Show full details for a channel",
	Long:  "Show full details for a channel",
}

func init() {
	channelCmd.AddCommand(channelInspectCmd)
}

func (r *runners) channelInspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appChan, _, err := r.api.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.ChannelAttrs(r.w, appChan); err != nil {
		return err
	}

	return nil
}
