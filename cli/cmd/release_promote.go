package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var releaseOptional bool
var releaseNotes string
var releaseVersion string

// releasePromoteCmd represents the releasePromote command
var releasePromoteCmd = &cobra.Command{
	Use:   "promote <sequence> <channelID>",
	Short: "Set the release for a channel",
	Long:  `replicated release promote <sequence> <channelID>`,
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
		cmd.Usage()
		os.Exit(1)
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %s", args[0])
	}
	chanID := args[1]

	if err := r.api.PromoteRelease(r.appID, seq, releaseVersion, releaseNotes, !releaseOptional, chanID); err != nil {
		return err
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", chanID, seq)
	r.w.Flush()

	return nil
}
