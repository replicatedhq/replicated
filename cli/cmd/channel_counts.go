package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

var channelCountsCmd = &cobra.Command{
	Use:   "counts CHANNEL_ID",
	Short: "Print channel license counts",
	Long:  "Print channel license counts",
}

func init() {
	channelCmd.AddCommand(channelCountsCmd)
}

func (r *runners) channelCounts(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appChan, _, err := r.api.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.LicenseCounts(r.w, appChan.LicenseCounts); err != nil {
		return err
	}

	return nil
}
