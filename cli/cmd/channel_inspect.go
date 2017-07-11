package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
)

// channelInspectCmd represents the channelInspect command
var channelInspectCmd = &cobra.Command{
	Use:   "inspect CHANNEL_ID",
	Short: "Show full details for a channel",
	Long:  "Show full details for a channel",
}

func init() {
	channelCmd.AddCommand(channelInspectCmd)
}

func (r *runners) channelInspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appChan, releases, err := r.api.GetChannel(r.appID, chanID)
	if err != nil {
		return err
	}

	if err = print.ChannelAttrs(r.w, appChan); err != nil {
		return err
	}

	if _, err = fmt.Fprint(r.w, "\nADOPTION\n"); err != nil {
		return err
	}
	if err = print.ChannelAdoption(r.w, &appChan.Adoption); err != nil {
		return err
	}

	if _, err = fmt.Fprint(r.w, "\nLICENSE_COUNTS\n"); err != nil {
		return err
	}
	if err = print.LicenseCounts(r.w, &appChan.LicenseCounts); err != nil {
		return err
	}

	if _, err = fmt.Fprint(r.w, "\nRELEASES\n"); err != nil {
		return err
	}
	if err = print.ChannelReleases(r.w, releases); err != nil {
		return err
	}

	return nil
}
