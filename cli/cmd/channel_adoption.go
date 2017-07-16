package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var channelAdoptionCmd = &cobra.Command{
	Use:   "adoption CHANNEL_ID",
	Short: "Print channel adoption statistics by license type",
	Long:  "Print channel adoption statistics by license type",
}

func init() {
	channelCmd.AddCommand(channelAdoptionCmd)
}

func (r *runners) channelAdoption(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appChan, _, err := r.api.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.ChannelAdoption(r.w, appChan.Adoption); err != nil {
		return err
	}

	return nil
}
