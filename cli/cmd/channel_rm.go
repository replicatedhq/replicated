package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// channelRmCmd represents the channelRm command
var channelRmCmd = &cobra.Command{
	Use:   "rm CHANNEL_ID",
	Short: "Remove (archive) a channel",
	Long:  "Remove (archive) a channel",
}

func init() {
	channelCmd.AddCommand(channelRmCmd)
}

func (r *runners) channelRemove(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	if err := r.platformAPI.ArchiveChannel(r.appID, chanID); err != nil {
		return err
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully archived\n", chanID)
	r.w.Flush()

	return nil
}
