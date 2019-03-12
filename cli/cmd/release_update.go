package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"
)

var updateReleaseYaml string

var releaseUpdateCmd = &cobra.Command{
	Use:   "update SEQUENCE",
	Short: "Updated a release's yaml config",
	Long:  "Updated a release's yaml config",
}

func init() {
	releaseCmd.AddCommand(releaseUpdateCmd)

	releaseUpdateCmd.Flags().StringVar(&updateReleaseYaml, "yaml", "", "The new YAML config for this release")
}

func (r *runners) releaseUpdate(cmd *cobra.Command, args []string) error {
	if updateReleaseYaml == "" {
		return fmt.Errorf("yaml is required")
	}
	if updateReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return err
		}
		updateReleaseYaml = string(bytes)
	}
	if len(args) < 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid release sequence: %s", args[0])
	}

	if err := r.platformAPI.UpdateRelease(r.appID, seq, updateReleaseYaml); err != nil {
		return fmt.Errorf("Failure setting new yaml config for release: %v", err)
	}

	// ignore the error since operation was successful
	fmt.Fprintf(r.w, "Release %d updated\n", seq)
	r.w.Flush()

	return nil
}
