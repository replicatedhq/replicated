package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var releaseOptional bool
var releaseNotes string
var releaseVersion string

// releasePromoteCmd represents the releasePromote command
var releasePromoteCmd = &cobra.Command{
	Use:   "promote SEQUENCE CHANNEL_ID",
	Short: "Set the release for a channel",
	Long: `Set the release for a channel

Example: replicated release promote 15 fe4901690971757689f022f7a460f9b2`,
}

func init() {
	releaseCmd.AddCommand(releasePromoteCmd)

	releasePromoteCmd.Flags().StringVar(&releaseNotes, "release-notes", "", "The **markdown** release notes")
	releasePromoteCmd.Flags().BoolVar(&releaseOptional, "optional", false, "If set, this release can be skipped")
	releasePromoteCmd.Flags().StringVar(&releaseVersion, "version", "", "A version label for the release in this channel")
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

	if err := r.platformAPI.PromoteRelease(r.appID, seq, releaseVersion, releaseNotes, !releaseOptional, chanID); err != nil {
		return err
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", chanID, seq)
	r.w.Flush()

	return nil
}
