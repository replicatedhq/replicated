package cmd

import (
	"fmt"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

var createReleaseYaml string

var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new release",
	Long: `Create a new release by providing YAML configuration for the next release in
your sequence.`,
}

func init() {
	releaseCmd.AddCommand(releaseCreateCmd)

	releaseCreateCmd.Flags().StringVar(&createReleaseYaml, "yaml", "", "The YAML config for this release")
}

func (r *runners) releaseCreate(cmd *cobra.Command, args []string) error {
	if createReleaseYaml == "" {
		return fmt.Errorf("yaml is required")
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

	return nil
}
