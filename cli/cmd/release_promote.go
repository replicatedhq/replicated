package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/client"
	"github.com/replicatedhq/replicated/pkg/logger"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleasePromote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:           "promote SEQUENCE CHANNEL_ID",
		Short:         "Set the release for a channel",
		Long:          `Set the release for a channel`,
		Example:       `replicated release promote 15 fe4901690971757689f022f7a460f9b2`,
		SilenceErrors: true, // this command uses custom error printing
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.releaseNotes, "release-notes", "", "The **markdown** release notes")
	cmd.Flags().BoolVar(&r.args.releaseOptional, "optional", false, "If set, this release can be skipped")
	cmd.Flags().BoolVar(&r.args.releaseRequired, "required", false, "If set, this release can't be skipped")
	cmd.Flags().StringVar(&r.args.releaseVersion, "version", "", "A version label for the release in this channel")
	cmd.Flags().BoolVar(&r.args.releasePromoteWaitForAirgap, "wait-for-airgap", false, "Wait for airgap bundle builds to complete (KOTS apps only)")
	cmd.Flags().DurationVar(&r.args.releasePromoteWaitForAirgapTimeout, "wait-for-airgap-timeout", 30*time.Minute, "Timeout for waiting on airgap bundle builds")

	cmd.RunE = r.releasePromote
}

var inFlightAirgapStates = map[string]bool{
	"pending":         true,
	"building":        true,
	"building_bundle": true,
	"metadata":        true,
	"unknown":         true,
}

// terminalFailureStates are the airgap-build outcomes that should fail the
// --wait-for-airgap waiter. "warn" is included because the airgap-builder uses
// it for builds that completed with non-fatal warnings vendors still need to
// see — silently treating it as success would reproduce the misleading-success
// bug this feature is meant to fix.
var terminalFailureStates = map[string]bool{
	"failed":               true,
	"failed_with_metadata": true,
	"cancelled":            true,
	"warn":                 true,
}

func (r *runners) releasePromote(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()

	if !r.hasApp() {
		return errors.New("no app specified")
	}

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

	promoteResp, err := r.api.PromoteRelease(r.appID, r.appType, seq, r.args.releaseVersion, r.args.releaseNotes, required, newID)
	if err != nil {
		return errors.Wrapf(err, "failed to promote release")
	}

	fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", channelName, seq)

	if r.appType == "kots" && r.args.releasePromoteWaitForAirgap {
		log := logger.NewLogger(r.w).SetIsTerminal(r.stdoutIsTTY)
		if err := r.waitForAirgapBuilds(promoteResp, r.args.releasePromoteWaitForAirgapTimeout, log); err != nil {
			return err
		}
	}

	r.w.Flush()

	return nil
}

func (r *runners) waitForAirgapBuilds(promoteResp *types.PromoteReleaseResponse, timeout time.Duration, log *logger.Logger) error {
	// The vandoor API always returns airgapBuilds[] for promoted channels (metadata
	// generation runs for every channel-release). This branch only fires when talking
	// to an older server that did not populate the field.
	if promoteResp != nil && len(promoteResp.AirgapBuilds) == 0 {
		log.ActionWithoutSpinner("No airgap build status reported by the API (older server, or no channels were promoted)")
		return nil
	}

	log.ActionWithSpinner("Waiting for airgap builds")

	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if time.Since(start) > timeout {
			log.FinishSpinnerWithError()
			return errors.New("timed out waiting for airgap builds")
		}

		allTerminal := true
		var failures []string
		for _, build := range promoteResp.AirgapBuilds {
			status, err := r.kotsAPI.GetAirgapBuildStatus(r.appID, build.ChannelID, build.ChannelSequence)
			if err != nil {
				log.FinishSpinnerWithError()
				return errors.Wrapf(err, "failed to get airgap build status for channel %s", build.ChannelName)
			}

			if inFlightAirgapStates[status.AirgapBuildStatus] {
				allTerminal = false
			} else if terminalFailureStates[status.AirgapBuildStatus] {
				failures = append(failures, fmt.Sprintf("channel %s (%s): %s", status.ChannelName, status.AirgapBuildStatus, status.AirgapBuildError))
			}
		}

		if allTerminal {
			if len(failures) > 0 {
				log.FinishSpinnerWithError()
				for _, f := range failures {
					log.ChildActionWithoutSpinner("%s", f)
				}
				return errors.New("one or more airgap builds failed")
			}
			log.FinishSpinner()
			log.ChildActionWithoutSpinner("Airgap builds complete")
			return nil
		}

		<-ticker.C
	}
}
