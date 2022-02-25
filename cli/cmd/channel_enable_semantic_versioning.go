package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelEnableSemanticVersioning(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "enable-semantic-versioning CHANNEL_ID",
		Short: "Enable semantic versioning for CHANNEL_ID",
		Long: `Enable semantic versioning for the CHANNEL_ID.

 Example:
 replicated channel enable-semantic-versioning CHANNEL_ID`,
	}
	cmd.Hidden = true // Not supported in KOTS
	parent.AddCommand(cmd)
	cmd.RunE = r.channelEnableSemanticVersioning
}

func (r *runners) channelEnableSemanticVersioning(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]
	if err := r.api.UpdateSemanticVersioningForChannel(r.appType, r.appID, chanID, true); err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Semver successfully enabled for channel %s\n", chanID)
	r.w.Flush()

	return nil
}
