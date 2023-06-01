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
	cmd.Flags().StringVar(&r.args.channelInspectOutputFormat, "output", "text", "The output format to use. One of: json|text (default: text)")

	cmd.RunE = r.channelInspect
}

func (r *runners) channelInspect(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel name or ID is required")
	}

	channelNameOrID := args[0]
	appChan, err := r.api.GetChannelByName(r.appID, r.appType, channelNameOrID)
	if err != nil {
		return err
	}

	if err = print.ChannelAttrs(r.args.channelInspectOutputFormat, r.w, r.appType, r.appSlug, appChan); err != nil {
		return err
	}

	return nil
}
