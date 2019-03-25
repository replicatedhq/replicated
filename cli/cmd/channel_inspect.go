package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
)

func (r *runners) InitChannelInspect(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "inspect CHANNEL_ID",
		Short: "Show full details for a channel",
		Long:  "Show full details for a channel",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.channelInspect
}

func (r *runners) channelInspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appChan, _, err := r.platformAPI.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.ChannelAttrs(r.w, appChan); err != nil {
		return err
	}

	return nil
}
