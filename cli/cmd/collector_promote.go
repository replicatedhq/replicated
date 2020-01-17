package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (r *runners) InitCollectorPromote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "promote SPEC_ID CHANNEL_ID",
		Short: "Set the collector for a channel",
		Long: `Set the collector for a channel

  Example: replicated collectors promote skd204095829040 fe4901690971757689f022f7a460f9b2`,
	}
	cmd.Hidden=true; // Not supported in KOTS 
	parent.AddCommand(cmd)

	cmd.RunE = r.collectorPromote
}

func (r *runners) collectorPromote(cmd *cobra.Command, args []string) error {
	// parse spec ID and channel ID positional arguments
	if len(args) != 2 {
		return errors.New("collector spec ID and channel ID are required")
	}
	specID := args[0]
	chanID := args[1]

	err := r.api.PromoteCollector(r.appID, specID, chanID)

	if err != nil {
		return err
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Collector %s successfully promoted to channel %s\n", specID, chanID)

	r.w.Flush()

	return nil

}
