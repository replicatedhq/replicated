package cmd

import (
	"encoding/json"
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
// --wait-for-airgap waiter. "warn" is intentionally NOT in this map — the
// bundle still exists, just with soft warnings about unresolvable image
// references; the waiter surfaces those messages but exits 0. See the
// dedicated "warn" branch in waitForAirgapBuilds.
var terminalFailureStates = map[string]bool{
	"failed":               true,
	"failed_with_metadata": true,
	"cancelled":            true,
}

func (r *runners) releasePromote(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		printIfError(cmd, err)
	}()
	// Flush the tabwriter on every return path. Without this, returning early
	// from waitForAirgapBuilds (e.g. on terminal failure) would discard the
	// "Channel successfully set" line and the per-channel failure messages
	// the waiter wrote, leaving the user with only the generic error output.
	defer r.w.Flush()

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

	if r.outputFormat == "json" {
		log := logger.NewLogger(r.w).SetIsTerminal(r.stdoutIsTTY)
		log.Silence()

		out := struct {
			Channel         string   `json:"channel"`
			ReleaseSequence int64    `json:"release_sequence"`
			VersionLabel    string   `json:"version_label"`
			Warnings        []string `json:"warnings,omitempty"`
		}{
			Channel:         newID,
			ReleaseSequence: seq,
			VersionLabel:    r.args.releaseVersion,
		}

		if promoteResp != nil && len(promoteResp.Warnings) > 0 {
			out.Warnings = promoteResp.Warnings
		}

		enc := json.NewEncoder(r.w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			return errors.Wrap(err, "encode json output")
		}
	} else {
		if len(promoteResp.Warnings) > 0 {
			for _, w := range promoteResp.Warnings {
				fmt.Fprintf(r.w, "Warning: %s\n", w)
			}
		} else {
			fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", channelName, seq)
		}
	}

	if r.appType == "kots" && r.args.releasePromoteWaitForAirgap {
		log := logger.NewLogger(r.w).SetIsTerminal(r.stdoutIsTTY)
		if r.outputFormat == "json" {
			log.Silence()
		}
		if err := r.waitForAirgapBuilds(promoteResp, r.args.releasePromoteWaitForAirgapTimeout, log); err != nil {
			return err
		}
	}

	return nil
}

func (r *runners) waitForAirgapBuilds(promoteResp *types.PromoteReleaseResponse, timeout time.Duration, log *logger.Logger) error {
	// Short-circuit before the range below if the response is nil or carries no
	// airgap builds — both would otherwise nil-pointer-deref or no-op the loop.
	if promoteResp == nil || len(promoteResp.AirgapBuilds) == 0 {
		log.ActionWithoutSpinner("No airgap builds reported for this promotion")
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
		var warnings []string
		for _, build := range promoteResp.AirgapBuilds {
			status, err := r.kotsAPI.GetAirgapBuildStatus(r.appID, build.ChannelID, build.ChannelSequence)
			if err != nil {
				// Log the error but continue polling other channels — one channel's
				// transient API failure shouldn't mask the status of the rest.
				log.ChildActionWithoutSpinner("Warning: could not check airgap status for channel %s: %v", build.ChannelName, err)
				allTerminal = false
				continue
			}

			if inFlightAirgapStates[status.AirgapBuildStatus] {
				allTerminal = false
			} else if terminalFailureStates[status.AirgapBuildStatus] {
				failures = append(failures, fmt.Sprintf("channel %s (%s): %s", status.ChannelName, status.AirgapBuildStatus, status.AirgapBuildError))
			} else if status.AirgapBuildStatus == "warn" {
				warnings = append(warnings, fmt.Sprintf("channel %s (warn): %s", status.ChannelName, status.AirgapBuildError))
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
			if len(warnings) > 0 {
				for _, w := range warnings {
					log.ChildActionWithoutSpinner("%s", w)
				}
			}
			log.ChildActionWithoutSpinner("Airgap builds complete")
			return nil
		}

		<-ticker.C
	}
}
