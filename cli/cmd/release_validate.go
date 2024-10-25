package cmd

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseTest(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "test SEQUENCE",
		Short: "Test the application release",
		Long:  "Test the application release",
	}
	parent.AddCommand(cmd)
	cmd.RunE = r.releaseTest
}

func (r *runners) releaseTest(command *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(args) != 1 {
		return errors.New("release sequence is required")
	}
	seq, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to parse sequence argument %q", args[0])
	}

	result, err := r.api.TestRelease(r.appID, r.appType, seq)
	if err != nil {
		return errors.Wrap(err, "test release")
	}

	fmt.Printf("Test results for release %#v", result)

	return nil

}
