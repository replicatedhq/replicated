package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createReleaseYaml string

var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new release",
	Long:  `Provide YAML configuration for the next release in your sequence.`,
}

func init() {
	releaseCmd.AddCommand(releaseCreateCmd)

	releaseCreateCmd.Flags().StringVar(&createReleaseYaml, "yaml", "", "The YAML config for this release")
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) error {
	// TODO can cobra do this?
	if createReleaseYaml == "" {
		return fmt.Errorf("yaml is required")
	}

	// API does not accept yaml in create operation, so first create then udpate
	release, err := r.api.CreateRelease(r.appID)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(r.w, "SEQUENCE: %d\n", release.Sequence); err != nil {
		return err
	}
	r.w.Flush()

	if err := r.api.UpdateRelease(r.appID, release.Sequence, createReleaseYaml); err != nil {
		return fmt.Errorf("Failure setting yaml config for release: %v", err)
	}

	return nil
}
