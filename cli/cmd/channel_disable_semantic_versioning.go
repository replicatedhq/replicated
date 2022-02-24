package cmd

import (
	"errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelDisableSemanticVersioning(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "disable-semantic-versioning CHANNEL_ID",
		Short: "Disable semantic versioning for CHANNEL_ID",
		Long: `Disable semantic versioning for the CHANNEL_ID.

 Example:
 replicated channel disable-semantic-versioning CHANNEL_ID`,
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)
	cmd.RunE = r.channelDisableSemanticVersioning
}

func (r *runners) channelDisableSemanticVersioning(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]
	if err := r.api.UpdateSemanticVersioningForChannel(r.appType, r.appID, chanID, false); err != nil {
		return err
	}

	return nil
}
