package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicatedhq/replicated/cli/print"
)

// channelInspectCmd represents the channelInspect command
var channelInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Show full details for a channel",
	Long:  `replicated channel inspect be52315888f23408e2e4dc9242d4cc2c`,
}

func init() {
	channelCmd.AddCommand(channelInspectCmd)
}

func (r *runners) channelInspect(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf(cmd.UsageString())
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
