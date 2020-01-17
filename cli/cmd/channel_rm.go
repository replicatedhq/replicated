package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitChannelRemove(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "rm CHANNEL_ID",
		Short: "Remove (archive) a channel",
		Long:  "Remove (archive) a channel",
	}
	cmd.Hidden=true; // Not supported in KOTS (ch #22646)
	parent.AddCommand(cmd)
	cmd.RunE = r.channelRemove
}

func (r *runners) channelRemove(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	if err := r.api.ArchiveChannel(r.appID, r.appType, chanID); err != nil {
		return err
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully archived\n", chanID)
	r.w.Flush()

	return nil
}
