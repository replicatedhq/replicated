package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitChannelRemove(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "rm CHANNEL_ID_OR_NAME",
		Aliases: []string{"delete"},
		Short:   "Remove (archive) a channel",
		Long:    "Remove (archive) a channel",
	}
	parent.AddCommand(cmd)
	cmd.RunE = r.channelRemove
}

func (r *runners) channelRemove(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(args) != 1 {
		return errors.New("channel name or ID is required")
	}

	channelNameOrID := args[0]
	channel, err := r.api.GetChannelByName(r.appID, r.appType, channelNameOrID)
	if err != nil {
		return err
	}

	if err := r.api.ArchiveChannel(r.appID, r.appType, channel.ID); err != nil {
		return err
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully archived\n", channelNameOrID)
	r.w.Flush()

	return nil
}
