package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorPromote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "promote NAME CHANNEL_ID",
		Short: "Set the collector for a channel",
		Long: `Set the collector for a channel

  Example: replicated collector promote test-collector fe4901690971757689f022f7a460f9b2`,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.collectorName, "name", "", "A name for the collector in this channel")

	cmd.RunE = r.collectorPromote
}

func (r *runners) collectorPromote(cmd *cobra.Command, args []string) error {
	// parse name and channel ID positional arguments
	if len(args) != 2 {
		return errors.New("collector name and channel ID are required")
	}
	name := args[0]
	chanID := args[1]

	if err := r.api.PromoteCollector(r.appID, name, chanID); err != nil {
		return err
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully set to collector %s\n", chanID, name)
	r.w.Flush()

	return nil
}
