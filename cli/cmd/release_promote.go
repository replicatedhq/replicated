package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (r *runners) InitReleasePromote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "promote SEQUENCE CHANNEL_ID",
		Short: "Set the release for a channel",
		Long: `Set the release for a channel

  Example: replicated release promote 15 fe4901690971757689f022f7a460f9b2`,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.releaseNotes, "release-notes", "", "The **markdown** release notes")
	cmd.Flags().BoolVar(&r.args.releaseOptional, "optional", false, "If set, this release can be skipped")
	cmd.Flags().StringVar(&r.args.releaseVersion, "version", "", "A version label for the release in this channel")

	cmd.RunE = r.releasePromote
}

func (r *runners) releasePromote(cmd *cobra.Command, args []string) error {
	// parse sequence and channel ID positional arguments
	if len(args) != 2 {
		return errors.New("releasese sequence and channel ID are required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %s", args[0])
	}
	chanID := args[1]

	if err := r.platformAPI.PromoteRelease(r.appID, seq, r.args.releaseVersion, r.args.releaseNotes, !r.args.releaseOptional, chanID); err != nil {
		return err
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", chanID, seq)
	r.w.Flush()

	return nil
}
