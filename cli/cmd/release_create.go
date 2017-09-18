package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

var createReleaseYaml string
var createReleasePromote string

var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new release",
	Long: `Create a new release by providing YAML configuration for the next release in
your sequence.`,
}

func init() {
	releaseCmd.AddCommand(releaseCreateCmd)

	releaseCreateCmd.Flags().StringVar(&createReleaseYaml, "yaml", "", "The YAML config for this release")
	releaseCreateCmd.Flags().StringVar(&createReleasePromote, "promote", "", "Channel name or id to promote this release to")
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) error {
	if createReleaseYaml == "" {
		return fmt.Errorf("yaml is required")
	}

	if createReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		createReleaseYaml = string(bytes)
	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if createReleasePromote != "" {
		channels, err := r.api.ListChannels(r.appID)
		if err != nil {
			return err
		}

		promoteChannelIDs := make([]string, 0)
		for _, c := range channels {
			if c.Id == createReleasePromote || c.Name == createReleasePromote {
				promoteChannelIDs = append(promoteChannelIDs, c.Id)
			}
		}

		if len(promoteChannelIDs) == 0 {
			return fmt.Errorf("Channel %q not found", createReleasePromote)
		}
		if len(promoteChannelIDs) > 1 {
			return fmt.Errorf("Channel %q is ambiguous. Please use channel ID", createReleasePromote)
		}
		promoteChanID = promoteChannelIDs[0]
	}

	opts := &client.ReleaseOptions{
		YAML: createReleaseYaml,
	}
	release, err := r.api.CreateRelease(r.appID, opts)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.w, "SEQUENCE: %d\n", release.Sequence); err != nil {
		return err
	}
	r.w.Flush()

	if promoteChanID != "" {
		if err := r.api.PromoteRelease(
			r.appID,
			release.Sequence,
			"",
			"",
			false,
			promoteChanID,
		); err != nil {
			return err
		}

		// ignore error since operation was successful
		fmt.Fprintf(r.w, "Channel %s successfully set to release %d\n", promoteChanID, release.Sequence)
		r.w.Flush()
	}

	return nil
}
