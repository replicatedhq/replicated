package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelAdoption(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "adoption CHANNEL_ID",
		Short: "Print channel adoption statistics by license type",
		Long:  "Print channel adoption statistics by license type",
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.channelAdoption
}

func (r *runners) channelAdoption(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("channel ID is required")
	}
	chanID := args[0]

	appType, err := r.api.GetAppType(r.appID)
	if err != nil {
		return err
	}

	if appType == "platform" {
		appChan, _, err := r.platformAPI.GetChannel(r.appID, chanID)
		if err != nil {
			return err
		}

		if err = print.ChannelAdoption(r.w, appChan.Adoption); err != nil {
			return err
		}

	} else if appType == "ship" {
		return errors.New("This feature is not supported for Ship applications.")
	} else if appType == "kots" {
		return errors.New("This feature is not supported for Kots applications.")
	}

	return nil
}
