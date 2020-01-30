package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelCounts(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "counts CHANNEL_ID",
		Short: "Print channel license counts",
		Long:  "Print channel license counts",
	}
	cmd.Hidden=true; // Not supported in KOTS 
	parent.AddCommand(cmd)
	cmd.RunE = r.channelCounts
}

func (r *runners) channelCounts(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	if r.appType == "platform" {
		appChan, _, err := r.api.GetChannel(r.appID, r.appType, chanID)
		if err != nil {
			return err
		}

		if err = print.LicenseCounts(r.stdoutWriter, appChan.LicenseCounts); err != nil {
			return err
		}
	} else if r.appType == "ship" {
		return errors.New("This feature is not supported for Ship applications.")
	} else if r.appType == "kots" {
		return errors.New("This feature is not supported for Kots applications.")
	}

	return nil
}
