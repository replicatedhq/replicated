package cmd

import (
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all channels in your app",
		Long:  "List all channels in your app",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.channelList
}

func (r *runners) channelList(cmd *cobra.Command, args []string) error {
	channels, err := r.api.ListChannels(r.appID, r.appType, "")
	if err != nil {
		return err
	}

	return print.Channels(r.w, channels)
}
