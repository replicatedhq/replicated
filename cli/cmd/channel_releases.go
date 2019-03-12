package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var channelReleasesCmd = &cobra.Command{
	Use:   "releases CHANNEL_ID",
	Short: "List all releases in a channel",
	Long:  "List all releases in a channel",
}

func init() {
	channelCmd.AddCommand(channelReleasesCmd)
}

func (r *runners) channelReleases(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	_, releases, err := r.platformAPI.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.ChannelReleases(r.w, releases); err != nil {
		return err
	}

	return nil
}
