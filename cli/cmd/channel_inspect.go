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
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.channelInspect
}

func (r *runners) channelInspect(_ *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(args) != 1 {
		return errors.New("channel name or ID is required")
	}

	channelNameOrID := args[0]
	appChan, err := r.api.GetChannelByName(r.appID, r.appType, channelNameOrID)
	if err != nil {
		return err
	}

	if err = print.ChannelAttrs(r.outputFormat, r.w, r.appType, r.appSlug, appChan); err != nil {
		return err
	}

	return nil
}
