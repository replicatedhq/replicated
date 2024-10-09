package cmd

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleasePromote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "promote SEQUENCE CHANNEL_ID",
		Short: "Set the release for a channel",
		Long: `Set the release for a channel

  Example: replicated release promote 15 fe4901690971757689f022f7a460f9b2`,
		SilenceErrors: false,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.releaseNotes, "release-notes", "", "The **markdown** release notes")
	cmd.Flags().BoolVar(&r.args.releaseOptional, "optional", false, "If set, this release can be skipped")
	cmd.Flags().BoolVar(&r.args.releaseRequired, "required", false, "If set, this release can't be skipped")
	cmd.Flags().StringVar(&r.args.releaseVersion, "version", "", "A version label for the release in this channel")

	cmd.RunE = r.releasePromote
}

func (r *runners) releasePromote(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceErrors = true // this command uses custom error printing

	defer func() {
		printIfError(cmd, err)
	}()

	// parse sequence and channel ID positional arguments
	if len(args) != 2 {
		return errors.New("release sequence and channel ID are required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %s", args[0])
	}
	channelName := args[1]
	newID := channelName

	// try to turn chanID into an actual id if it was a channel name
	opts := client.GetOrCreateChannelOptions{
		AppID:          r.appID,
		AppType:        r.appType,
		NameOrID:       channelName,
		CreateIfAbsent: false,
	}
	channelID, err := r.api.GetOrCreateChannelByName(opts)
	if err != nil {
		return errors.Wrapf(err, "unable to get channel ID from name")
	}
	newID = channelID.ID

	required := false
	if r.appType == "platform" {
		required = !r.args.releaseOptional
	} else if r.appType == "kots" {
		required = r.args.releaseRequired
	}

	if err = r.api.PromoteRelease(r.appID, r.appType, seq, r.args.releaseVersion, r.args.releaseNotes, required, newID); err != nil {
		return errors.Wrapf(err, "failed to promote release")
	}

	// ignore error since operation was successful
	fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", channelName, seq)
	r.w.Flush()

	return nil
}
