package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelReleases(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "releases CHANNEL_ID",
		Short: "List all releases in a channel",
		Long:  "List all releases in a channel",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.channelReleases
}

func (r *runners) channelReleases(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	if r.appType == "platform" {

		_, releases, err := r.api.GetChannel(r.appID, r.appType, chanID)
		if err != nil {
			return err
		}

		if err = print.ChannelReleases(r.w, releases); err != nil {
			return err
		}
	} else if r.appType == "ship" {
		return errors.New("This feature is not supported for Ship applications.")
	} else if r.appType == "kots" {
		return errors.New("This feature is not supported for Kots applications.")
	}

	return nil
}
